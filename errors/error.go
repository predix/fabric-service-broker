package errors

const ErrAsyncResponse = `
{
  "error": "AsyncRequired",
  "description": "This service plan requires client support for asynchronous service operations."
}
`
const ErrNetworksUnavailable = `
{
  "error": "NetworkUnavailable",
  "description": "No networks available for deployments"
}
`
const ErrManifestGeneration = `
{
  "error": "ManifestGeneration",
  "description": "Unable to generate manifest for deployment"
}
`
const ErrHttpRequest = `
{
  "error": "HttpRequestCreate",
  "description": "Unable to create an http request"
}
`
const ErrBoshConnect = `
{
  "error": "BoshConnect",
  "description": "Unable to connect to Bosh"
}
`
const ErrDBSave = `
{
  "error": "DBSave",
  "description": "Unable to save to DB"
}
`

const ErrDBDelete = `
{
  "error": "DBDelete",
  "description": "Unable to delete from DB"
}
`

const ErrDBRead = `
{
  "error": "DBRead",
  "description": "Unable to read from DB"
}
`

const ErrBoshInvalidResponse = `
{
  "error": "BoshInvalidResponse",
  "description": "Invalid response from Bosh"
}
`

const ErrResourceAlreadyExists = `
{
  "error": "ResourceAlreadyExists",
  "description": "Resource already exists"
}
`

const ErrProvisionInFlight = `
{
  "error": "ProvisionInFlight",
  "description": "Service instance is still being deployed"
}
`

const ErrBindingsExist = `
{
  "error": "BindingExist",
  "description": "Service instance cannot be deleted as bindings exist"
}
`
