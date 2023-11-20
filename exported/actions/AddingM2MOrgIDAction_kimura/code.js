/**
* Handler that will be called during the execution of a Client Credentials exchange.
*
* @param {Event} event - Details about client credentials grant request.
* @param {CredentialsExchangeAPI} api - Interface whose methods can be used to change the behavior of client credentials grant.
*/
exports.onExecuteCredentialsExchange = async (event, api) => {

  try {
    if(event.client.metadata && event.client.metadata.org_id){
      let org_id = event.client.metadata.org_id;
      console.log('Fetching client metadata succeeded. m2m.org_id: ' + org_id + ' will be set in the access token');
      api.accessToken.setCustomClaim("m2m.org_id", org_id);
    }
  } catch (e) {
    console.log('Fetching client metadata failed. reason: ' + e);
  }
};
