package authentication

import (
	nex "github.com/PretendoNetwork/nex-go"
	"github.com/PretendoNetwork/nex-protocols-go/authentication"
)

func requestTicket(err error, client *nex.Client, callID uint32, userPID uint32, targetPID uint32) {
	encryptedTicket, errorCode := generateTicket(userPID, targetPID)

	rmcResponse := nex.NewRMCResponse(authentication.ProtocolID, callID)

	// TODO:
	// If the source or target pid is invalid, the %retval% field is set to Core::AccessDenied and the ticket is empty.
	if errorCode != 0 {
		rmcResponse.SetError(errorCode)
	} else {
		rmcResponseStream := nex.NewStreamOut(commonAuthenticationProtocol.server)

		rmcResponseStream.WriteResult(nex.NewResultSuccess(nex.Errors.Core.Unknown))
		rmcResponseStream.WriteBuffer(encryptedTicket)

		rmcResponseBody := rmcResponseStream.Bytes()

		rmcResponse.SetSuccess(authentication.MethodRequestTicket, rmcResponseBody)
	}

	rmcResponseBytes := rmcResponse.Bytes()

	var responsePacket nex.PacketInterface

	if commonAuthenticationProtocol.server.PRUDPVersion() == 0 {
		responsePacket, _ = nex.NewPacketV0(client, nil)
		responsePacket.SetVersion(0)
	} else {
		responsePacket, _ = nex.NewPacketV1(client, nil)
		responsePacket.SetVersion(1)
	}

	responsePacket.SetSource(0xA1)
	responsePacket.SetDestination(0xAF)
	responsePacket.SetType(nex.DataPacket)
	responsePacket.SetPayload(rmcResponseBytes)

	responsePacket.AddFlag(nex.FlagNeedsAck)
	responsePacket.AddFlag(nex.FlagReliable)

	commonAuthenticationProtocol.server.Send(responsePacket)
}
