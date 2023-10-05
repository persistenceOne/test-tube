use persistence_std::shim::Any;
use persistence_std::types::cosmos::base::v1beta1::Coin;
use persistence_std::types::cosmos::gov::v1::{MsgSubmitProposal, MsgSubmitProposalResponse, MsgVoteResponse, QueryProposalRequest, QueryProposalResponse, QueryParamsRequest, QueryParamsResponse, MsgVote, VoteOption};
use test_tube_x::{fn_execute, fn_query, Account, RunnerExecuteResult, SigningAccount};

use test_tube_x::module::Module;
use test_tube_x::runner::Runner;

use crate::PersistenceTestApp;

pub struct Gov<'a, R: Runner<'a>> {
    runner: &'a R,
}

impl<'a, R: Runner<'a>> Module<'a, R> for Gov<'a, R> {
    fn new(runner: &'a R) -> Self {
        Self { runner }
    }
}

impl<'a, R> Gov<'a, R>
where
    R: Runner<'a>,
{
    fn_execute! {
        pub submit_proposal: MsgSubmitProposal["/cosmos.gov.v1.MsgSubmitProposal"] => MsgSubmitProposalResponse
    }

    fn_execute! {
        pub vote: MsgVote["/cosmos.gov.v1.MsgVote"] => MsgVoteResponse
    }

    fn_query! {
        pub query_proposal ["/cosmos.gov.v1.Query/Proposal"]: QueryProposalRequest => QueryProposalResponse
    }

    fn_query! {
        pub query_params ["/cosmos.gov.v1.Query/Params"]: QueryParamsRequest => QueryParamsResponse
    }

    pub fn submit_executable_proposal(
        &self,
        title: String,
        metadata: String,
        summary: String,
        proposer: String,
        initial_deposit: Vec<cosmwasm_std::Coin>,
        messages: Vec<Any>,
        signer: &SigningAccount,
    ) -> RunnerExecuteResult<MsgSubmitProposalResponse> {
        self.submit_proposal(
            MsgSubmitProposal {
                // content: Some(Any {
                //     type_url: msg_type_url,
                //     value: msg
                //         .to_bytes()
                //         .map_err(|e| RunnerError::EncodeError(e.into()))?,
                // }),
                initial_deposit: initial_deposit
                    .into_iter()
                    .map(|coin| Coin {
                        denom: coin.denom,
                        amount: coin.amount.to_string(),
                    })
                    .collect(),
                proposer,
                messages,
                metadata,
                title,
                summary,
            },
            signer,
        )
    }
}

/// Extension for Gov module
/// It has ability to access to `OsmosisTestApp` which is more specific than `Runner`
pub struct GovWithAppAccess<'a> {
    gov: Gov<'a, PersistenceTestApp>,
    app: &'a PersistenceTestApp,
}

impl<'a> GovWithAppAccess<'a> {
    pub fn new(app: &'a PersistenceTestApp) -> Self {
        Self {
            gov: Gov::new(app),
            app,
        }
    }

    pub fn to_gov(&self) -> &Gov<'a, PersistenceTestApp> {
        &self.gov
    }

    pub fn propose_and_execute<M: prost::Message>(
        &self,
        msgs: Vec<Any>,
        proposer: String,
        signer: &SigningAccount,
    ) -> RunnerExecuteResult<MsgSubmitProposalResponse> {
        // query deposit params
        let params = self.gov.query_params(&QueryParamsRequest {
            params_type: "deposit".to_string(),
        })?;

        let min_deposit = params
            .params
            .expect("params must exist")
            .min_deposit;

        // submit proposal
        let submit_proposal_res = self.gov.submit_proposal(
            MsgSubmitProposal {
                messages: msgs,
                initial_deposit: min_deposit,
                proposer,
                metadata: "test metadata".to_string(),
                title: "test title".to_string(),
                summary: "test summary".to_string(),
            },
            signer,
        )?;

        let proposal_id = submit_proposal_res.data.proposal_id;

        // get validator to vote yes for proposal
        let val = self.app.get_first_validator_signing_account()?;
        self.gov
            .vote(
                MsgVote {
                    proposal_id,
                    voter: val.address(),
                    option: VoteOption::Yes.into(),
                    metadata: "test metadata".to_string(),
                },
                &val,
            )
            .unwrap();

        // query params
        let params = self.gov.query_params(&QueryParamsRequest {
            params_type: "voting".to_string(),
        })?;

        // get voting period
        let voting_period = params
            .params
            .expect("params must exist")
            .voting_period
            .expect("voting period must exist");

        // increase time to pass voting period
        self.app.increase_time(voting_period.seconds as u64 + 1);

        Ok(submit_proposal_res)
    }
}

#[cfg(test)]
mod tests {

    // #[test]
    // fn test_passing_and_execute_proposal() {
    //     let app = PersistenceTestApp::default();
    //     let gov = GovWithAppAccess::new(&app);

    //     let proposer = app
    //         .init_account(&[cosmwasm_std::Coin::new(1000000000000000000, "uosmo")])
    //         .unwrap();

    //     // query code id 1 should error since it has not been stored
    //     let err = app
    //         .query::<_, QueryCodeResponse>(
    //             "/cosmwasm.wasm.v1.Query/Code",
    //             &QueryCodeRequest { code_id: 1 },
    //         )
    //         .unwrap_err();

    //     assert_eq!(
    //         err,
    //         RunnerError::QueryError {
    //             msg: "not found".to_string()
    //         }
    //     );

    //     // store code through proposal
    //     let wasm_byte_code = std::fs::read("./test_artifacts/cw1_whitelist.wasm").unwrap();
    //     let res = gov
    //         .propose_and_execute(
    //             StoreCodeProposal::TYPE_URL.to_string(),
    //             StoreCodeProposal {
    //                 title: String::from("test"),
    //                 description: String::from("test"),
    //                 run_as: proposer.address(),
    //                 wasm_byte_code: wasm_byte_code.clone(),
    //                 instantiate_permission: None,
    //                 unpin_code: false,
    //                 source: String::new(),
    //                 builder: String::new(),
    //                 code_hash: Vec::new(),
    //             },
    //             proposer.address(),
    //             false,
    //             &proposer,
    //         )
    //         .unwrap();

    //     assert_eq!(res.data.proposal_id, 1);

    //     // query code id 1 should find the code after proposal is executed
    //     let QueryCodeResponse { code_info, data } = app
    //         .query(
    //             "/cosmwasm.wasm.v1.Query/Code",
    //             &QueryCodeRequest { code_id: 1 },
    //         )
    //         .unwrap();

    //     assert_eq!(code_info.unwrap().creator, proposer.address());
    //     assert_eq!(data, wasm_byte_code);
    // }

    // #[test]
    // fn test_cosmwasmpool_proposal() {
    //     let app = PersistenceTestApp::default();
    //     let gov = GovWithAppAccess::new(&app);

    //     let proposer = app
    //         .init_account(&[cosmwasm_std::Coin::new(1000000000000000000, "uosmo")])
    //         .unwrap();

    //     // upload cosmwasm pool code and whitelist through proposal
    //     let wasm_byte_code = std::fs::read("./test_artifacts/transmuter.wasm").unwrap();
    //     let res = gov
    //         .propose_and_execute(
    //             UploadCosmWasmPoolCodeAndWhiteListProposal::TYPE_URL.to_string(),
    //             UploadCosmWasmPoolCodeAndWhiteListProposal {
    //                 title: String::from("test"),
    //                 description: String::from("test"),
    //                 wasm_byte_code,
    //             },
    //             proposer.address(),
    //             false,
    //             &proposer,
    //         )
    //         .unwrap();

    //     assert_eq!(res.data.proposal_id, 1);
    // }
}
