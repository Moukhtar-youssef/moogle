extern crate redis;
use redis::Commands;
use std::env;
use std::{thread, time};
use std::process::Command;

fn main() {
    // Get environment variable
    let redis_host = env::var("REDIS_HOST").expect("REDIS_HOST environment variable not set");
    let redis_port = env::var("REDIS_PORT").unwrap_or_else(|_| "6379".to_string());
    let redis_password = env::var("REDIS_PASSWORD").unwrap_or_else(|_| "".to_string());
    let redis_db : u8 = env::var("REDIS_DB").unwrap_or("0".to_string()).parse().expect("REDIS_DB must be a valid integer");

    // Build the Redis URL
    let mut redis_url = format!("redis://{}:{}", redis_host, redis_port);

    // Add password if available
    if redis_password.is_empty() {
        redis_url = format!("redis://:{}@{}:{}", redis_password, redis_host, redis_port);
    }

    // Add the DB number
    redis_url = format!("{}/{}", redis_url, redis_db);

    println!("Connecting to Redis at: {}...", redis_url);
    // Connect to redis
    let client = redis::Client::open(redis_url).expect("Invalid Redis URL");
    let mut con = client.get_connection().expect("Failed to connect to Redis");

    // Get command line arguments for max_spiders and max_indexers
    let args: Vec<String> = env::args().collect();

    let max_spiders = if args.len() > 1 {
        args[1].parse::<usize>().unwrap_or(2)
    } else {
        2
    };

    let max_indexers = if args.len() > 2 {
        args[2].parse::<usize>().unwrap_or(10)
    } else {
        10
    };

    loop {
        // Get the length of the url_queue
        let url_queue_length: usize = con.llen("url_queue").expect("Failed to get queue length");
        // Get the length of the indexer_queue
        let indexer_queue_length: usize = con.llen("indexer_queue").expect("Failed to get queue length");
        // Print the length of the url_queue and the indexer_queue
        println!("|--------------------------------------|");
        println!("| Current url_queue length: {}", url_queue_length);
        println!("| Current indexer_queue length: {}", indexer_queue_length);
        println!("|--------------------------------------|");
        // Scale spiders
        let desired_spiders = {
            if url_queue_length < 5000 {
                1
            } else {
                std::cmp::min(
                    max_spiders,
                    (1.0 + (url_queue_length as f64 / 20000.0).sqrt()).min(10.0).floor() as usize
                )
            }
        };


        let desired_spiders = std::cmp::min(desired_spiders, max_spiders);
        let current_spiders = get_current_spiders();
        println!("| Current spiders: {}", current_spiders);
        println!("|\tDesired spider count: {}", desired_spiders);
        scale_spiders(desired_spiders);

        // Scale indexers
        let desired_indexers = {
            if indexer_queue_length < 500 {
                1
            } else {
                std::cmp::min(
                    max_indexers, 
                    (1.5 * (indexer_queue_length as f64 / 1000.0)).ceil() as usize
                )
            }
        };

        let desired_indexers = std::cmp::min(desired_indexers, max_indexers);
        let current_indexers = get_current_indexers();
        println!("| Current indexers: {}", current_indexers);
        println!("|\tDesired indexer count: {}", desired_indexers);
        scale_indexers(desired_indexers);

        println!("|--------------------------------------|");
        println!("");

        // Sleep
        thread::sleep(time::Duration::from_millis(5000));
    }
}

fn scale_spiders(desired_spiders: usize) {
    let current_spiders = get_current_spiders();

    if current_spiders != desired_spiders {
        println!("Scaling spiders to: {}", desired_spiders);
        Command::new("docker")
            .arg("compose")
            .arg("-f")
            .arg("../spider/docker-compose.yml")  // Specify the indexer service compose file
            .arg("up")
            .arg("--scale")
            .arg(format!("spider-service={}", desired_spiders))
            .arg("-d")
            .status()
            .expect("Failed to scale spiders");
    }
}

fn get_current_spiders() -> usize {
    let output = Command::new("docker")
        .arg("compose")
        .arg("-f")
        .arg("../spider/docker-compose.yml")  // Specify the indexer service compose file
        .arg("ps")
        .output()
        .expect("Failed to get Docker Compose services");

    let output_str = String::from_utf8_lossy(&output.stdout);

    output_str.split('\n')
        .filter(|line| line.contains("spider-service"))
        .count()
}

fn scale_indexers(desired_indexers: usize) {
    let current_indexers = get_current_indexers();

    if current_indexers != desired_indexers {
        println!("Scaling indexers to: {}", desired_indexers);
        Command::new("docker")
            .arg("compose")
            .arg("-f")
            .arg("../indexer/docker-compose.yml")  // Specify the indexer service compose file
            .arg("up")
            .arg("--scale")
            .arg(format!("indexer-service={}", desired_indexers))
            .arg("-d")
            .status()
            .expect("Failed to scale indexers");
    }
}

fn get_current_indexers() -> usize {
    let output = Command::new("docker")
        .arg("compose")
        .arg("-f")
        .arg("../indexer/docker-compose.yml")  // Specify the indexer service compose file
        .arg("ps")
        .output()
        .expect("Failed to get Docker Compose services");

    let output_str = String::from_utf8_lossy(&output.stdout);

    output_str.split('\n')
        .filter(|line| line.contains("indexer-service"))
        .count()
}

