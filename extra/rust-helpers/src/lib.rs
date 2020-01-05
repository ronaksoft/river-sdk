#![allow(dead_code, unused)]
#![feature(fixed_size_array)]
extern crate aead;
extern crate aes_gcm;
extern crate primes;
extern crate sha2;


use std::vec;

use aead::{Aead, NewAead};
use aead::generic_array::{
    GenericArray,
    typenum::{U0, U12, U16, U8},
};
use aes_gcm::{Aes256Gcm, AesGcm};
use sha2::{Digest, Sha512};
use sha2::digest::Reset;
use std::array::FixedSizeArray;

#[cfg(test)]
mod tests {
    use crate::{gen_message_key, split_pq, encrypt, decrypt};

    #[test]
    fn check_split_pq() {
        let pq: u64 = 273366826352239;
//        println!("Result is: {:?}", split_pq(pq))
    }

    #[test]
    fn check_gen_message_key() {
        let dh_key: [u8; 100] = [0; 100];
        let mut dh_key: Vec<u8> = Vec::from(&dh_key[..]);
        dh_key.extend_from_slice("1234567890123456789012345678901234567890".as_bytes());
//        println!("{}", dh_key.len());

        let body = "Hello It is Ehsan";
        let res = gen_message_key(dh_key.split_off(100).as_slice(), Vec::from(body));
//        println!("{:?}", res.as_slice());
    }

    #[test]
    fn check_encrypt() {
        let dh_key: [u8; 256] = [10; 256];
        let body = "Hi This is Me";
        let msg_key = gen_message_key(&dh_key[100..140], body.into());
        let cipher = encrypt( &dh_key, body.into()).expect("could not encrypt");
        let plain = decrypt(&dh_key, msg_key.as_slice(), body.into());
        println!("{:?}  {:?}", body, cipher);
        match plain {
            Ok(plain) => {
                println!("{:?} -> {:?}", cipher, plain);
            },
            Err(err) => {println!("{:?}", err)}
        }
    }
}

fn split_pq(pq: u64) -> (u64, u64) {
    let f = primes::factors(pq);
    (f[0], f[1])
}

fn gen_message_key(key: &[u8], plain: Vec<u8>) -> Vec<u8> {
    let mut hasher: Sha512 = Sha512::new();
    let mut v: Vec<u8> = Vec::from(key);
    v.append(&mut plain.clone());
    hasher.input(v.as_slice());
    hasher.result().to_vec().split_off(32)
}

fn encrypt(dh_key: &[u8; 256], plain: Vec<u8>) -> Result<Vec<u8>, aead::Error> {
    // Message Key is: _Sha512(DHKey[100:140], InternalHeader, Payload)[32:64]
    let mut msg_key: Vec<u8> = gen_message_key(&dh_key[100..140], plain.clone());
    // AES IV: _Sha512 (DHKey[180:220], MessageKey)[:32]
    let mut iv: Vec<u8> = Vec::from(&dh_key[180..220]);
    iv.extend_from_slice(msg_key.as_slice());
    let mut hasher: Sha512 = Sha512::new();
    hasher.input(iv.as_slice());
    let aes_iv = hasher.result().to_vec();
    let aes_iv = aes_iv.split_at(12);


    // AES KEY: _Sha512 (MessageKey, DHKey[170:210])[:32]
    let mut hasher: Sha512 = Sha512::new();
    let mut key: Vec<u8> = Vec::from(msg_key.as_slice());
    key.extend_from_slice(&dh_key[170..210]);
    hasher.input(key.as_slice());
    let aes_key = hasher.result().to_vec();
    let aes_key = aes_key.split_at(32);

    let aead :Aes256Gcm = AesGcm::new(GenericArray::from_slice(aes_key.0).clone());
    let nonce = GenericArray::from_slice(aes_iv.0); // 96-bits; unique per message

    aead.encrypt(nonce, plain.clone().as_ref())
}

fn decrypt(dh_key: &[u8; 256], msg_key:&[u8], cipher: Vec<u8>) -> Result<Vec<u8>, aead::Error> {
    // AES IV: _Sha512 (DHKey[180:220], MessageKey)[:32]
    let mut iv: Vec<u8> = Vec::from(&dh_key[180..220]);
    iv.extend(msg_key.iter());
    let mut hasher: Sha512 = Sha512::new();
    hasher.input(iv.as_slice());
    let aes_iv = hasher.result().to_vec();
    let aes_iv = aes_iv.split_at(12);

    // AES KEY: _Sha512 (MessageKey, DHKey[170:210])[:32]
    let mut hasher: Sha512 = Sha512::new();
    let mut key: Vec<u8> = Vec::new();
    key.extend_from_slice(msg_key);
    key.extend_from_slice(&dh_key[170..210]);
    hasher.input(key.as_slice());
    let aes_key = hasher.result().to_vec();
    let aes_key = aes_key.split_at(32);

    let aead :Aes256Gcm = AesGcm::new(GenericArray::from_slice(aes_key.0).clone());
    let nonce = GenericArray::from_slice(aes_iv.0); // 96-bits; unique per message


    aead.decrypt(nonce, cipher.as_slice())
}


