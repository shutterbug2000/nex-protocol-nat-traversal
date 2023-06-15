package nattraversal

import (
	"strconv"

	nex "github.com/PretendoNetwork/nex-go"
	nat_traversal "github.com/PretendoNetwork/nex-protocols-go/nat-traversal"
)

func reportNatProperties(err error, client *nex.Client, callID uint32, natm uint32, natf uint32, rtt uint32) {
	missingHandler := false
	if GetConnectionUrlsHandler == nil {
		logger.Warning("Missing GetConnectionUrlsHandler!")
		missingHandler = true
	}
	if ReplaceConnectionUrlHandler == nil {
		logger.Warning("Missing ReplaceConnectionUrlHandler!")
		missingHandler = true
	}
	if missingHandler {
		return
	}
	stationUrlsStrings := GetConnectionUrlsHandler(client.ConnectionID())
	stationUrls := make([]nex.StationURL, len(stationUrlsStrings))
	pid := strconv.FormatUint(uint64(client.PID()), 10)
	rvcid := strconv.FormatUint(uint64(client.ConnectionID()), 10)

	for i := 0; i < len(stationUrlsStrings); i++ {
		stationUrls[i] = *nex.NewStationURL(stationUrlsStrings[i])
		if stationUrls[i].Type() == "3" {
			natm_s := strconv.FormatUint(uint64(natm), 10)
			natf_s := strconv.FormatUint(uint64(natf), 10)
			stationUrls[i].SetNatm(natm_s)
			stationUrls[i].SetNatf(natf_s)
		}
		stationUrls[i].SetPID(pid)
		stationUrls[i].SetRVCID(rvcid)
		ReplaceConnectionUrlHandler(client.ConnectionID(), stationUrlsStrings[i], stationUrls[i].EncodeToString())
	}

	rmcResponse := nex.NewRMCResponse(nat_traversal.ProtocolID, callID)
	rmcResponse.SetSuccess(nat_traversal.MethodReportNATProperties, nil)

	rmcResponseBytes := rmcResponse.Bytes()

	var responsePacket nex.PacketInterface

	if server.PRUDPVersion() == 0 {
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
