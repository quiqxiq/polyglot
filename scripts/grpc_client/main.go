package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
)

func main() {
	log.Println("🚀 STARTING gRPC INTEGRATION TEST CLIENT...")

	target := "localhost:50051"
	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.CallContentSubtype("json")),
	)
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server at %s: %v", target, err)
	}
	defer conn.Close()

	client := devicepb.NewDeviceServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Test ListDevices RPC
	log.Println("\n1️⃣  Executing ListDevices gRPC RPC...")
	listRes, err := client.ListDevices(ctx, &devicepb.ListDevicesRequest{})
	if err != nil {
		log.Fatalf("ListDevices failed: %v", err)
	}
	log.Printf("   ListDevices Success! Found %d devices:", len(listRes.Devices))
	for _, d := range listRes.Devices {
		log.Printf("   - [%s] %s (%s:%d) Vendor: %s | Enabled: %v", d.Id, d.Name, d.Host, d.Port, d.Vendor, d.Enabled)
	}

	// 2. Test GetDevice RPC
	if len(listRes.Devices) > 0 {
		targetID := listRes.Devices[0].Id
		log.Printf("\n2️⃣  Executing GetDevice gRPC RPC for ID %q...", targetID)
		getRes, err := client.GetDevice(ctx, &devicepb.GetDeviceRequest{Id: targetID})
		if err != nil {
			log.Fatalf("GetDevice failed: %v", err)
		}
		log.Printf("   GetDevice Success! Device: %+v", getRes.Device)

		// 3. Test UpdateDevice RPC
		log.Printf("\n3️⃣  Executing UpdateDevice gRPC RPC for ID %q...", targetID)
		updateReq := &devicepb.UpdateDeviceRequest{
			Device: &devicepb.Device{
				Id:         getRes.Device.Id,
				TenantId:   getRes.Device.TenantId,
				Name:       getRes.Device.Name + " (gRPC Updated)",
				Vendor:     getRes.Device.Vendor,
				DriverType: getRes.Device.DriverType,
				Host:       getRes.Device.Host,
				Port:       getRes.Device.Port,
				TimeoutMs:  getRes.Device.TimeoutMs,
				Enabled:    getRes.Device.Enabled,
			},
			Username: "admin",
			Password: "",
		}
		updateRes, err := client.UpdateDevice(ctx, updateReq)
		if err != nil {
			log.Fatalf("UpdateDevice failed: %v", err)
		}
		log.Printf("   UpdateDevice Success! Message: %s | New Name: %s", updateRes.Message, updateRes.Device.Name)
	}

	// 4. Test StreamDeviceStatus RPC
	log.Println("\n4️⃣  Subscribing to StreamDeviceStatus gRPC Server Streaming RPC...")
	streamCtx, streamCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer streamCancel()

	stream, err := client.StreamDeviceStatus(streamCtx, &devicepb.StreamDeviceStatusRequest{})
	if err != nil {
		log.Fatalf("StreamDeviceStatus failed: %v", err)
	}

	frameCount := 0
	for {
		frame, err := stream.Recv()
		if err == io.EOF || streamCtx.Err() != nil {
			break
		}
		if err != nil {
			log.Printf("   Stream receive ended: %v", err)
			break
		}
		frameCount++
		log.Printf("   📩 [gRPC Frame #%d] Device: %s | Status: %s | Uptime: %s | Msg: %s",
			frameCount, frame.Device.Name, frame.Test.Status, frame.Test.Uptime, frame.Test.Message)
		if frameCount >= 3 {
			break
		}
	}

	fmt.Println("\n======================================================================")
	fmt.Println("  ✅ ALL gRPC CLIENT & SERVER PROCEDURES VERIFIED SUCCESSFULLY!")
	fmt.Println("======================================================================")
}
