#![allow(dead_code, unused)]
extern crate num_bigint;
extern crate num_traits;
extern crate rand;

use std::cmp::Ordering::{Equal, Greater, Less};
use std::ops::{Add, BitAnd, BitAndAssign, Div, Shr, Sub};

use num_bigint::{BigInt, RandBigInt, Sign};
use num_traits::{One, Zero, zero};
use rand::{Rng, thread_rng};

#[cfg(test)]
mod tests {
    use crate::split_pq;

    use std::cmp::Ordering::{Equal, Greater, Less};
    use std::ops::{Add, BitAnd, BitAndAssign, Div, Shr, Sub};
    use num_bigint::{BigInt, RandBigInt, Sign};
    use num_traits::{One, Zero, zero};
    use rand::{Rng, thread_rng};


    #[test]
    fn check_split_pq() {
        let pq = BigInt::new(Sign::Plus, vec![2, 1]);
        println!("Result is: {:?}", split_pq(pq))
    }
}

fn split_pq(pq: BigInt) -> (BigInt, BigInt) {
    let value0: BigInt = Zero::zero();
    let value1: BigInt = One::one();
    let value15 = BigInt::new(Sign::Plus, vec![1, 5]);
    let value17 = BigInt::new(Sign::Plus, vec![1, 7]);
    let rnd_max = BigInt::from(1i64 << 63);

    let mut rng = rand::thread_rng();
    let what = pq.clone();
    let rnd = rng.gen::<i64>();
    let mut g = BigInt::new(Sign::Plus, vec![0]);
    let mut i = 0;

    while !(g.cmp(&value1) == Greater && g.cmp(&what) == Less) {

        let q: BigInt = rng.gen_range(value0, rnd_max);
        q.bitand(&value15);
        q.bitand(&value17);


        let qm = q.mod_floor(&what);
        break;
//        let mut x = rng.gen_bigint_range(Zero::zero(), &rndMax);
//        let what_next = what.clone().sub(&value1);
//
//        let mut y = BigInt::from(x.mod_floor(&what_next).add(&value1));
//        x = BigInt::from(y);
//        let lim = 1 << (i + 18);
//        let mut j = 1;
//        let mut flag = true;
//        while j < lim && flag {
//            let mut a = BigInt::from(&x);
//            let mut b = BigInt::from(&x);
//            let mut c = BigInt::from(&q);
//            while b.cmp(&value0) == Greater {
//                let mut b2: BigInt = zero();
//                if b2.bitand_assign(&value1).cmp(&value0) == Greater {
//                    c = c.add(&what);
//                    if c.cmp(&what) == Greater || c.cmp(&what) == Equal {
//                        c = c.sub(&what);
//                    }
//                    a = a.add(&a);
//                    if a.cmp(&what) == Greater || c.cmp(&what) == Equal {
//                        a = a.sub(&what);
//                    }
//                    b = b.shr(1);
//                }
//                x = c.clone();
//                let mut z: BigInt = zero();
//                if x.cmp(&y) == Less {
//                    z = what.add(&x).sub(&y);
//                } else {
//                    z = x.sub(&y);
//                }
//                g = z.gcd(&what);
//                if j.bitand(j - 1) == 0 {
//                    y = x;
//                }
//                j = j + 1;
//                if g.cmp(&value1) != 0 {
//                    flag = false;
//                }
//            }
//            i = i + 1;
//        }
    }
    let mut p1 = g;
    let mut p2: BigInt = what.div(&g);

    (p1, p2)
}

