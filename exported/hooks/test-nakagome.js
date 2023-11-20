/**
@param {object} client - information about the client
@param {string} client.name - name of client
@param {string} client.id - client id
@param {string} client.tenant - Auth0 tenant name
@param {object} client.metadata - client metadata
@param {array|undefined} scope - array of strings representing the scope claim or undefined
@param {string} audience - token's audience claim
@param {object} context - additional authorization context
@param {object} context.webtask - webtask context
@param {function} cb - function (error, accessTokenClaims)
*/
module.exports = function(client, scope, audience, context, cb) {
  var access_token = {};
  access_token.scope = scope;
  console.log({client})
  console.log({context})
  // Modify scopes or add extra claims
  access_token['m2m.org_id'] = context.body.org_id;
  // access_token.scope.push('extra');

  // Deny the token and respond with an OAuth2 error response
  // if (denyExchange) {
  //   // To return an HTTP 400 with { "error": "invalid_scope", "error_description": "Not authorized for this scope." }
  //   return cb(new InvalidScopeError('Not authorized for this scope.'));
  //
  //   // To return an HTTP 400 with { "error": "invalid_request", "error_description": "Not a valid request." }
  //   return cb(new InvalidRequestError('Not a valid request.'));
  //
  //   // To return an HTTP 500 with { "error": "server_error", "error_description": "A server error occurred." }
  //   return cb(new ServerError('A server error occurred.'));
  // }

  cb(null, access_token);
};
