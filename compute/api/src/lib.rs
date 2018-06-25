extern crate futures;
extern crate grpcio;
extern crate protobuf;

extern crate ekiden_consensus_api;

mod generated;

use ekiden_consensus_api as consensus;

pub use generated::computation_group::*;
pub use generated::computation_group_grpc::*;
pub use generated::web3::*;
pub use generated::web3_grpc::*;
