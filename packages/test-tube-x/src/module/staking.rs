#![cfg(feature = "bank")]

use crate::{fn_execute, fn_query};
use persistence_std::types::cosmos::staking::v1beta1::{
    MsgBeginRedelegate, MsgBeginRedelegateResponse, MsgCreateValidator, MsgCreateValidatorResponse,
    MsgDelegate, MsgDelegateResponse, MsgEditValidator, MsgEditValidatorResponse,
    MsgUnbondValidator, QueryDelegationRequest, QueryDelegationResponse,
    QueryDelegatorDelegationsRequest, QueryDelegatorDelegationsResponse,
    QueryDelegatorUnbondingDelegationsRequest, QueryDelegatorUnbondingDelegationsResponse,
    QueryParamsRequest, QueryParamsResponse, QueryPoolRequest, QueryPoolResponse,
    QueryRedelegationsRequest, QueryRedelegationsResponse, QueryUnbondingDelegationRequest,
    QueryUnbondingDelegationResponse, QueryValidatorRequest, QueryValidatorResponse,
    QueryValidatorsRequest, QueryValidatorsResponse,
};

use crate::module::Module;
use crate::runner::Runner;

pub struct Staking<'a, R: Runner<'a>> {
    runner: &'a R,
}

impl<'a, R: Runner<'a>> Module<'a, R> for Staking<'a, R> {
    fn new(runner: &'a R) -> Self {
        Self { runner }
    }
}

impl<'a, R> Staking<'a, R>
where
    R: Runner<'a>,
{
    fn_execute! {
        pub delegate: MsgDelegate["/cosmos.staking.v1beta1.MsgDelegate"] => MsgDelegateResponse
    }

    fn_execute! {
        pub create_validator: MsgCreateValidator["/cosmos.staking.v1beta1.MsgCreateValidator"] => MsgCreateValidatorResponse
    }

    fn_execute! {
        pub edit_validator: MsgEditValidator["/cosmos.staking.v1beta1.MsgEditValidator"] => MsgEditValidatorResponse
    }

    fn_execute! {
        pub redelegate: MsgBeginRedelegate["/cosmos.staking.v1beta1.MsgBeginRedelegate"] => MsgBeginRedelegateResponse
    }

    fn_execute! {
        pub unbond: MsgUnbondValidator["/cosmos.staking.v1beta1.MsgUnbondValidator"] => MsgUnbondValidator
    }

    fn_query! {
        pub query_delegation ["/cosmos.bank.v1beta1.Query/Balance"]: QueryDelegationRequest => QueryDelegationResponse
    }

    fn_query! {
        pub query_delegations ["/cosmos.bank.v1beta1.Query/AllBalances"]: QueryDelegatorDelegationsRequest => QueryDelegatorDelegationsResponse
    }

    fn_query! {
        pub query_unbonding_delegation ["/cosmos.bank.v1beta1.Query/Balance"]: QueryUnbondingDelegationRequest => QueryUnbondingDelegationResponse
    }

    fn_query! {
        pub query_unbonding_delegations ["/cosmos.bank.v1beta1.Query/AllBalances"]: QueryDelegatorUnbondingDelegationsRequest => QueryDelegatorUnbondingDelegationsResponse
    }

    fn_query! {
        pub query_redelegations ["/cosmos.staking.v1beta1.Query/Redelegations"]: QueryRedelegationsRequest => QueryRedelegationsResponse
    }

    fn_query! {
        pub query_validator ["/cosmos.bank.v1beta1.Query/Balance"]: QueryValidatorRequest => QueryValidatorResponse
    }

    fn_query! {
        pub query_validators ["/cosmos.bank.v1beta1.Query/AllBalances"]: QueryValidatorsRequest => QueryValidatorsResponse
    }

    fn_query! {
        pub query_params ["/cosmos.bank.v1beta1.Query/AllBalances"]: QueryParamsRequest => QueryParamsResponse
    }

    fn_query! {
        pub query_pool ["/cosmos.bank.v1beta1.Query/AllBalances"]: QueryPoolRequest => QueryPoolResponse
    }
}
