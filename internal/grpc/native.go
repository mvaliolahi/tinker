package grpc

import (
        "context"
        "fmt"
        "time"

        "google.golang.org/grpc"
        "google.golang.org/grpc/credentials/insecure"
        "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

const grpcDialTimeout = 10 * time.Second

// NativeClient provides gRPC operations using native Go libraries,
// eliminating the need for the grpcurl binary for one-shot commands.
type NativeClient struct {
        addr       string
        protoDir   string
        reflection bool
}

// NewNativeClient creates a native gRPC client.
func NewNativeClient(addr, protoDir string, reflection bool) *NativeClient {
        return &NativeClient{
                addr:       addr,
                protoDir:   protoDir,
                reflection: reflection,
        }
}

// ListServices lists all gRPC services using server reflection.
//
//nolint:staticcheck // grpc_reflection_v1alpha is deprecated but still the standard reflection API
func (n *NativeClient) ListServices() (string, error) {
        if !n.reflection {
                return "", fmt.Errorf("server reflection is disabled — enable it in tinker.toml or install grpcurl for proto-file mode")
        }

        conn, err := n.dial()
        if err != nil {
                return "", err
        }
        defer conn.Close()

        client := grpc_reflection_v1alpha.NewServerReflectionClient(conn)
        stream, err := client.ServerReflectionInfo(context.Background())
        if err != nil {
                return "", fmt.Errorf("reflection stream: %w", err)
        }
        defer func() { _ = stream.CloseSend() }()

        // Request list of services
        if err := stream.Send(&grpc_reflection_v1alpha.ServerReflectionRequest{
                MessageRequest: &grpc_reflection_v1alpha.ServerReflectionRequest_ListServices{},
        }); err != nil {
                return "", fmt.Errorf("sending reflection request: %w", err)
        }

        resp, err := stream.Recv()
        if err != nil {
                return "", fmt.Errorf("receiving reflection response: %w", err)
        }

        services := resp.GetListServicesResponse()
        if services == nil {
                return "", fmt.Errorf("unexpected reflection response type")
        }

        var result string
        for _, svc := range services.GetService() {
                result += svc.GetName() + "\n"
        }

        return result, nil
}

// Describe returns the file descriptor for a service using server reflection.
//
//nolint:staticcheck // grpc_reflection_v1alpha is deprecated but still the standard reflection API
func (n *NativeClient) Describe(service string) (string, error) {
        if !n.reflection {
                return "", fmt.Errorf("server reflection is disabled — enable it in tinker.toml or install grpcurl for proto-file mode")
        }

        conn, err := n.dial()
        if err != nil {
                return "", err
        }
        defer conn.Close()

        client := grpc_reflection_v1alpha.NewServerReflectionClient(conn)
        stream, err := client.ServerReflectionInfo(context.Background())
        if err != nil {
                return "", fmt.Errorf("reflection stream: %w", err)
        }
        defer func() { _ = stream.CloseSend() }()

        // Request the file containing the service
        if err := stream.Send(&grpc_reflection_v1alpha.ServerReflectionRequest{
                MessageRequest: &grpc_reflection_v1alpha.ServerReflectionRequest_FileContainingSymbol{
                        FileContainingSymbol: service,
                },
        }); err != nil {
                return "", fmt.Errorf("sending reflection request: %w", err)
        }

        resp, err := stream.Recv()
        if err != nil {
                return "", fmt.Errorf("receiving reflection response: %w", err)
        }

        fdResp := resp.GetFileDescriptorResponse()
        if fdResp == nil {
                return "", fmt.Errorf("service %q not found via reflection", service)
        }

        var result string
        for _, fdBytes := range fdResp.GetFileDescriptorProto() {
                desc := formatFileDescriptor(fdBytes, service)
                if desc != "" {
                        result += desc + "\n"
                }
        }

        if result == "" {
                return "", fmt.Errorf("could not describe service %q", service)
        }

        return result, nil
}

// Call invokes a gRPC method using server reflection to discover the schema.
// For simple JSON-in/JSON-out calls, this uses raw grpcurl encoding.
// Falls back to grpcurl CLI for methods requiring complex proto serialization.
func (n *NativeClient) Call(_, _ string) (string, error) {
        // Native gRPC call with proto serialization requires the full proto
        // descriptor, which we don't have at runtime without proto files.
        // For reflection-enabled servers, we attempt to use grpcurl-style
        // dynamic invocation, but this is limited without the proto codec.
        //
        // For now, native Call requires the grpcurl CLI for proper proto
        // serialization. The List and Describe commands work natively.
        return "", fmt.Errorf("native gRPC call requires proto codec — install grpcurl for full call support, or use reflection-enabled services with tinker grpc list/describe")
}

// dial establishes a gRPC connection with appropriate timeout.
//
//nolint:staticcheck // DialContext and WithBlock are deprecated in grpc v1.x but still functional
func (n *NativeClient) dial() (*grpc.ClientConn, error) {
        ctx, cancel := context.WithTimeout(context.Background(), grpcDialTimeout)
        defer cancel()

        opts := []grpc.DialOption{
                grpc.WithTransportCredentials(insecure.NewCredentials()),
                grpc.WithBlock(),
        }

        conn, err := grpc.DialContext(ctx, n.addr, opts...)
        if err != nil {
                return nil, fmt.Errorf("connecting to %s: %w", n.addr, err)
        }
        return conn, nil
}

// formatFileDescriptor extracts human-readable service description from a file descriptor.
func formatFileDescriptor(fdBytes []byte, targetService string) string {
        // Parse the file descriptor proto to extract service information
        fd, err := parseFileDescriptor(fdBytes)
        if err != nil {
                return fmt.Sprintf("// file descriptor (%d bytes)", len(fdBytes))
        }

        var result string
        for _, svc := range fd.Services {
                if targetService != "" && svc.Name != targetService {
                        continue
                }
                result += fmt.Sprintf("service %s {\n", svc.Name)
                for _, m := range svc.Methods {
                        result += fmt.Sprintf("  rpc %s(% s) returns (%s) {}\n",
                                m.Name, m.InputType, m.OutputType)
                }
                result += "}\n"
        }

        return result
}

// rawFileDescriptor is a simplified parsed representation.
type rawFileDescriptor struct {
        Package  string
        Services []rawService
}

type rawService struct {
        Name    string
        Methods []rawMethod
}

type rawMethod struct {
        Name       string
        InputType  string
        OutputType string
}

// parseFileDescriptor does a minimal parse of a protobuf FileDescriptorProto
// to extract service names and method signatures.
func parseFileDescriptor(data []byte) (*rawFileDescriptor, error) {
        fd := &rawFileDescriptor{}

        // Minimal protobuf wire format parsing for FileDescriptorProto
        // Field 2 (package) = string, field 7 (service) = repeated message
        // This is a lightweight parser that handles the common cases

        i := 0
        for i < len(data) {
                fieldNum, wireType, n := decodeVarint(data[i:])
                if n == 0 {
                        break
                }
                i += n

                switch {
                case fieldNum == 2 && wireType == 2: // package
                        str, nn := decodeString(data[i:])
                        fd.Package = str
                        i += nn

                case fieldNum == 7 && wireType == 2: // service
                        svcData, nn := decodeBytes(data[i:])
                        svc := parseService(svcData)
                        if svc != nil {
                                fd.Services = append(fd.Services, *svc)
                        }
                        i += nn

                default:
                        i = skipField(data[i:], wireType)
                }
        }

        return fd, nil
}

func parseService(data []byte) *rawService {
        svc := &rawService{}
        i := 0
        for i < len(data) {
                fieldNum, wireType, n := decodeVarint(data[i:])
                if n == 0 {
                        break
                }
                i += n

                switch {
                case fieldNum == 1 && wireType == 2: // name
                        str, nn := decodeString(data[i:])
                        svc.Name = str
                        i += nn

                case fieldNum == 2 && wireType == 2: // method
                        methodData, nn := decodeBytes(data[i:])
                        method := parseMethod(methodData)
                        if method != nil {
                                svc.Methods = append(svc.Methods, *method)
                        }
                        i += nn

                default:
                        i = skipField(data[i:], wireType)
                }
        }
        return svc
}

func parseMethod(data []byte) *rawMethod {
        m := &rawMethod{}
        i := 0
        for i < len(data) {
                fieldNum, wireType, n := decodeVarint(data[i:])
                if n == 0 {
                        break
                }
                i += n

                switch {
                case fieldNum == 1 && wireType == 2: // name
                        str, nn := decodeString(data[i:])
                        m.Name = str
                        i += nn

                case fieldNum == 2 && wireType == 2: // input_type
                        str, nn := decodeString(data[i:])
                        m.InputType = str
                        i += nn

                case fieldNum == 3 && wireType == 2: // output_type
                        str, nn := decodeString(data[i:])
                        m.OutputType = str
                        i += nn

                default:
                        i = skipField(data[i:], wireType)
                }
        }
        return m
}

// Protobuf wire format helpers

//nolint:gosec // integer overflow is acceptable in protobuf varint parsing
func decodeVarint(data []byte) (fieldNum, wireType, n int) {
        var x uint64
        for n = 0; n < len(data); n++ {
                b := data[n]
                x |= uint64(b&0x7F) << (7 * uint(n))
                if b&0x80 == 0 {
                        n++
                        break
                }
        }
        fieldNum = int(x >> 3)
        wireType = int(x & 0x7)
        return
}

//nolint:gosec // integer overflow is acceptable in protobuf varint parsing
func decodeString(data []byte) (string, int) {
        length, n := decodeRawVarint(data)
        if n+int(length) > len(data) {
                return string(data[n:]), n + len(data[n:])
        }
        return string(data[n : n+int(length)]), n + int(length)
}

//nolint:gosec // integer overflow is acceptable in protobuf varint parsing
func decodeBytes(data []byte) ([]byte, int) {
        length, n := decodeRawVarint(data)
        if n+int(length) > len(data) {
                return data[n:], n + len(data[n:])
        }
        return data[n : n+int(length)], n + int(length)
}

//nolint:gosec // integer overflow is acceptable in protobuf varint parsing
func decodeRawVarint(data []byte) (uint64, int) {
        var x uint64
        var n int
        for n = 0; n < len(data); n++ {
                b := data[n]
                x |= uint64(b&0x7F) << (7 * uint(n))
                if b&0x80 == 0 {
                        n++
                        return x, n
                }
        }
        return x, n
}

func skipField(data []byte, wireType int) int {
        switch wireType {
        case 0: // varint
                for i := 0; i < len(data); i++ {
                        if data[i]&0x80 == 0 {
                                return i + 1
                        }
                }
                return len(data)
        case 1: // 64-bit
                if len(data) < 8 {
                        return len(data)
                }
                return 8
        case 2: // length-delimited
                length, n := decodeRawVarint(data)
                total := n + int(length) //nolint:gosec // integer overflow acceptable in protobuf parsing
                if total > len(data) {
                        return len(data)
                }
                return total
        case 5: // 32-bit
                if len(data) < 4 {
                        return len(data)
                }
                return 4
        default:
                return len(data)
        }
}
