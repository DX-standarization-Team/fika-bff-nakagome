/**
* Handler that will be called during the execution of a Client Credentials exchange.
*
* @param {Event} event - Details about client credentials grant request.
* @param {CredentialsExchangeAPI} api - Interface whose methods can be used to change the behavior of client credentials grant.
*/
exports.onExecuteCredentialsExchange = async (event, api) => {
  // client_idでメタデータ検索
  const clientId = event.request.body.client_id;
  const clientSecret = event.request.body.client_secret;
  console.log('clientId: ' + clientId)
  console.log('clientSecret: ' + clientSecret)
  const ManagementClient = require('auth0').ManagementClient;
  const management = new ManagementClient({
      domain: 'dev-kjqwuq76z8suldgw.us.auth0.com',
      clientId: clientId,
      clientSecret: clientSecret,
      scope: "read:clients",
      audience: 'https://dev-kjqwuq76z8suldgw.us.auth0.com/api/v2/',
      tokenProvider: {
        enableCache: true,
        cacheTTLInSeconds: 10
      }
  });
  try{
    const res = await management.getClient({ client_id: clientId })
    console.log(res);
    if(res.client_metadata && res.client_metadata.org_id){ // ManagementAPIは設定がないから条件式追加
      const org_id = res.client_metadata.org_id;
      console.log('Fetching client metadata succeeded. m2m.org_id: ' + org_id + ' will be set in the access token');
      // m2m.org_idをカスタムクレームにセット
      api.accessToken.setCustomClaim("m2m.org_id", org_id);  
    }
  }catch(err){
    console.log(err)
    throw new Error('Fetching client metadata failed: ' + err);
  }
};