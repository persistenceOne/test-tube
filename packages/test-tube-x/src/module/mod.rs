use crate::runner::Runner;

pub mod bank;
pub mod wasm;
pub mod staking;

#[cfg(feature = "bank")]
pub use bank::Bank;

#[cfg(feature = "wasm")]
pub use wasm::Wasm;

#[cfg(feature = "staking")]
pub use staking::Staking;

#[macro_use]
pub mod macros;

pub trait Module<'a, R: Runner<'a>> {
    fn new(runner: &'a R) -> Self;
}
