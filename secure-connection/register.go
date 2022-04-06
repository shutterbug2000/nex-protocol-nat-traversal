package secureconnection

import (
	"fmt"
	"strconv"

	nex "github.com/PretendoNetwork/nex-go"
	nexproto "github.com/PretendoNetwork/nex-protocols-go"
)

func register(err error, client *nex.Client, callID uint32, stationUrls []*nex.StationURL) {
	missingHandler := false
	if AddConnectionHandler == nil {
		fmt.Println("Secure::Register missing AddConnectionHandler!")
		missingHandler = true
	}
	if UpdateConnectionHandler == nil {
		fmt.Println("Secure::Register missing UpdateConnectionHandler!")
		missingHandler = true
	}
	if DoesConnectionExistHandler == nil {
		fmt.Println("Secure::Register missing DoesConnectionExistHandler!")
		missingHandler = true
	}
	if missingHandler {
		return
	}
	localStation := stationUrls[0]
	localStationURL := localStation.EncodeToString()
	connectionID := uint32(server.ConnectionIDCounter().Increment())
	client.SetConnectionID(connectionID)
	client.SetLocalStationUrl(localStationURL)

	address := client.Address().IP.String()
	port := strconv.Itoa(client.Address().Port)
	natf := "0"
	natm := "0"
	type_ := "3"

	localStation.SetAddress(&address)
	localStation.SetPort(&port)
	localStation.SetNatf(&natf)
	localStation.SetNatm(&natm)
	localStation.SetType(&type_)

	globalStationURL := localStation.EncodeToString()

	if !DoesConnectionExistHandler(connectionID) {
		fmt.Println(localStationURL)
		AddConnectionHandler(connectionID, []string{localStationURL, globalStationURL}, address, port)
	} else {
		UpdateConnectionHandler(connectionID, []string{localStationURL, globalStationURL}, address, port)
	}

	rmcResponseStream := nex.NewStreamOut(server)

	rmcResponseStream.WriteUInt32LE(0x10001) // Success
	rmcResponseStream.WriteUInt32LE(connectionID)
	rmcResponseStream.WriteString(globalStationURL)

	rmcResponseBody := rmcResponseStream.Bytes()

	// Build response packet
	rmcResponse := nex.NewRMCResponse(nexproto.SecureProtocolID, callID)
	rmcResponse.SetSuccess(nexproto.SecureMethodRegister, rmcResponseBody)

	rmcResponseBytes := rmcResponse.Bytes()

	var responsePacket nex.PacketInterface

	if server.PrudpVersion() == 0 {
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

	server.Send(responsePacket)
}