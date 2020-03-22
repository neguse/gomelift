# dev memo

## TODO

### HealthCheck nonblocking

### Implement not implemented functions

### Connection management e.g. reconnecting, connection establish handling

### More correct error handling. Remove debug log.

## GameLiftServerAPI

user <-> sdk API surface

holds GameLiftServerState

### InitSDK()

Create GameLiftServerState and InitializeNetworking()
create two sio::client.
one for AuxProxyMessageSender
one for Network::Network

### ProcessReady()

GameLiftServerState->ProcessReady()
bind callbacks
getAuxProxySender->ProcessReady()
wake healthcheck thread
per 60sec ReportHealth() // TODO

### ProcessEnding()

GameLiftServerState->ProcessEnding()
getAuxProxySender->ProcessEnding()

### ActivateGameSession()

GameLiftServerState->ActivateGameSession()
getAuxProxySender->ActivateGameSession()

### TerminateGameSession()

GameLiftServerState->TerminateGameSession()
getAuxProxySender->TerminateGameSession()

### StartMatchBackfill()

GameLiftServerState->BackfillMatchmaking()

### StopMatchBackfill()

GameLiftServerState->StopMatchmaking()

### UpdatePlayerSessionCreationPolicy()

GameLiftServerState->UpdatePlayerSessionCreationPolicy()

### GetGameSessionId()

return false if !IsProcessReady()

GameLiftServerState->GetGameSessionId()

### GetTerminationTime()

GameLiftServerState->GetTerminationTime()

### AcceptPlayerSession()

return false if !IsProcessReady()

GameLiftServerState->AcceptPlayerSession()

### RemovePlayerSession()

return false if !IsProcessReady()

GameLiftServerState->RemovePlayerSession()

### Destroy()

GameLiftServerState::DestroyInstance()

### GetInstanceCertificate()

GameLiftServerState->GetInstanceCertificate()

## AuxProxyMessageSender

