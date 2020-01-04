#![allow(dead_code, unused)]
use std::vec;

fn main() {
    let a :[u8; 100] = [0; 100];
    let mut v : Vec<u8> = Vec::new();
    v.extend_from_slice(a.iter().as_slice());
    println!("{:?}", v);
}