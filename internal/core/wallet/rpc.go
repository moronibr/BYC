package wallet

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RPCServer represents the RPC server for hardware wallet operations
type RPCServer struct {
	manager *HardwareWalletManager
	limiter *rate.Limiter
	mu      sync.RWMutex
	// Store active sessions
	sessions map[string]*Session
}

// Session represents an authenticated RPC session
type Session struct {
	Token      string
	ExpiresAt  time.Time
	DeviceID   string
	LastAccess time.Time
}

// RPCMethod represents an RPC method
type RPCMethod struct {
	Name         string
	Handler      func(ctx context.Context, params json.RawMessage) (interface{}, error)
	AuthRequired bool
	RateLimit    rate.Limit
}

// RPCRequest represents an RPC request
type RPCRequest struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
	ID     interface{}     `json:"id"`
}

// RPCResponse represents an RPC response
type RPCResponse struct {
	Result interface{} `json:"result,omitempty"`
	Error  *RPCError   `json:"error,omitempty"`
	ID     interface{} `json:"id"`
}

// RPCError represents an RPC error
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewRPCServer creates a new RPC server
func NewRPCServer(manager *HardwareWalletManager) *RPCServer {
	return &RPCServer{
		manager:  manager,
		limiter:  rate.NewLimiter(rate.Limit(100), 100), // 100 requests per second
		sessions: make(map[string]*Session),
	}
}

// RegisterMethods registers all RPC methods
func (s *RPCServer) RegisterMethods() map[string]RPCMethod {
	methods := map[string]RPCMethod{
		"connect_device": {
			Name:         "connect_device",
			Handler:      s.handleConnectDevice,
			AuthRequired: false,
			RateLimit:    10,
		},
		"disconnect_device": {
			Name:         "disconnect_device",
			Handler:      s.handleDisconnectDevice,
			AuthRequired: true,
			RateLimit:    10,
		},
		"get_device_info": {
			Name:         "get_device_info",
			Handler:      s.handleGetDeviceInfo,
			AuthRequired: true,
			RateLimit:    20,
		},
		"initialize_device": {
			Name:         "initialize_device",
			Handler:      s.handleInitializeDevice,
			AuthRequired: true,
			RateLimit:    5,
		},
		"update_firmware": {
			Name:         "update_firmware",
			Handler:      s.handleUpdateFirmware,
			AuthRequired: true,
			RateLimit:    2,
		},
		"set_pin": {
			Name:         "set_pin",
			Handler:      s.handleSetPin,
			AuthRequired: true,
			RateLimit:    5,
		},
		"change_pin": {
			Name:         "change_pin",
			Handler:      s.handleChangePin,
			AuthRequired: true,
			RateLimit:    5,
		},
		"unlock_device": {
			Name:         "unlock_device",
			Handler:      s.handleUnlockDevice,
			AuthRequired: true,
			RateLimit:    10,
		},
		"lock_device": {
			Name:         "lock_device",
			Handler:      s.handleLockDevice,
			AuthRequired: true,
			RateLimit:    10,
		},
		"get_firmware_version": {
			Name:         "get_firmware_version",
			Handler:      s.handleGetFirmwareVersion,
			AuthRequired: true,
			RateLimit:    20,
		},
		"get_firmware_update_status": {
			Name:         "get_firmware_update_status",
			Handler:      s.handleGetFirmwareUpdateStatus,
			AuthRequired: true,
			RateLimit:    20,
		},
		"get_battery_level": {
			Name:         "get_battery_level",
			Handler:      s.handleGetBatteryLevel,
			AuthRequired: true,
			RateLimit:    20,
		},
		"set_passphrase": {
			Name:         "set_passphrase",
			Handler:      s.handleSetPassphrase,
			AuthRequired: true,
			RateLimit:    5,
		},
		"remove_passphrase": {
			Name:         "remove_passphrase",
			Handler:      s.handleRemovePassphrase,
			AuthRequired: true,
			RateLimit:    5,
		},
		"get_backup_status": {
			Name:         "get_backup_status",
			Handler:      s.handleGetBackupStatus,
			AuthRequired: true,
			RateLimit:    20,
		},
		"create_backup": {
			Name:         "create_backup",
			Handler:      s.handleCreateBackup,
			AuthRequired: true,
			RateLimit:    2,
		},
		"restore_backup": {
			Name:         "restore_backup",
			Handler:      s.handleRestoreBackup,
			AuthRequired: true,
			RateLimit:    2,
		},
		"get_connected_devices": {
			Name:         "get_connected_devices",
			Handler:      s.handleGetConnectedDevices,
			AuthRequired: true,
			RateLimit:    20,
		},
		"get_device_status": {
			Name:         "get_device_status",
			Handler:      s.handleGetDeviceStatus,
			AuthRequired: true,
			RateLimit:    20,
		},
		"get_device_capabilities": {
			Name:         "get_device_capabilities",
			Handler:      s.handleGetDeviceCapabilities,
			AuthRequired: true,
			RateLimit:    20,
		},
		"get_device_metrics": {
			Name:         "get_device_metrics",
			Handler:      s.handleGetDeviceMetrics,
			AuthRequired: true,
			RateLimit:    20,
		},
		"get_device_logs": {
			Name:         "get_device_logs",
			Handler:      s.handleGetDeviceLogs,
			AuthRequired: true,
			RateLimit:    20,
		},
		"get_device_events": {
			Name:         "get_device_events",
			Handler:      s.handleGetDeviceEvents,
			AuthRequired: true,
			RateLimit:    20,
		},
		"export_public_key": {
			Name:         "export_public_key",
			Handler:      s.handleExportPublicKey,
			AuthRequired: true,
			RateLimit:    20,
		},
		"verify_message": {
			Name:         "verify_message",
			Handler:      s.handleVerifyMessage,
			AuthRequired: true,
			RateLimit:    20,
		},
		"sign_message": {
			Name:         "sign_message",
			Handler:      s.handleSignMessage,
			AuthRequired: true,
			RateLimit:    20,
		},
		"get_xpub": {
			Name:         "get_xpub",
			Handler:      s.handleGetXPub,
			AuthRequired: true,
			RateLimit:    20,
		},
		"get_address": {
			Name:         "get_address",
			Handler:      s.handleGetAddress,
			AuthRequired: true,
			RateLimit:    20,
		},
		"get_address_info": {
			Name:         "get_address_info",
			Handler:      s.handleGetAddressInfo,
			AuthRequired: true,
			RateLimit:    20,
		},
	}

	return methods
}

// ServeHTTP handles HTTP requests
func (s *RPCServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check rate limit
	if !s.limiter.Allow() {
		s.sendError(w, &RPCError{
			Code:    -32000,
			Message: "rate limit exceeded",
		}, nil)
		return
	}

	// Parse request
	var req RPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, &RPCError{
			Code:    -32700,
			Message: "parse error",
		}, nil)
		return
	}

	// Get method
	method, exists := s.RegisterMethods()[req.Method]
	if !exists {
		s.sendError(w, &RPCError{
			Code:    -32601,
			Message: "method not found",
		}, req.ID)
		return
	}

	// Check authentication
	if method.AuthRequired {
		token := r.Header.Get("Authorization")
		if !s.validateSession(token) {
			s.sendError(w, &RPCError{
				Code:    -32001,
				Message: "unauthorized",
			}, req.ID)
			return
		}
	}

	// Handle request
	ctx := r.Context()
	result, err := method.Handler(ctx, req.Params)
	if err != nil {
		s.sendError(w, &RPCError{
			Code:    -32000,
			Message: err.Error(),
		}, req.ID)
		return
	}

	// Send response
	s.sendResponse(w, result, req.ID)
}

// sendResponse sends an RPC response
func (s *RPCServer) sendResponse(w http.ResponseWriter, result interface{}, id interface{}) {
	response := RPCResponse{
		Result: result,
		ID:     id,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// sendError sends an RPC error response
func (s *RPCServer) sendError(w http.ResponseWriter, err *RPCError, id interface{}) {
	response := RPCResponse{
		Error: err,
		ID:    id,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// validateSession validates a session token
func (s *RPCServer) validateSession(token string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[token]
	if !exists {
		return false
	}

	if time.Now().After(session.ExpiresAt) {
		delete(s.sessions, token)
		return false
	}

	session.LastAccess = time.Now()
	return true
}

// RPC method handlers

func (s *RPCServer) handleConnectDevice(ctx context.Context, params json.RawMessage) (interface{}, error) {
	if err := s.manager.ConnectToDevices(); err != nil {
		return nil, fmt.Errorf("failed to connect to devices: %w", err)
	}
	return "connected", nil
}

func (s *RPCServer) handleDisconnectDevice(ctx context.Context, params json.RawMessage) (interface{}, error) {
	if err := s.manager.DisconnectFromDevices(); err != nil {
		return nil, fmt.Errorf("failed to disconnect from devices: %w", err)
	}
	return "disconnected", nil
}

func (s *RPCServer) handleGetDeviceInfo(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		DeviceID string `json:"device_id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	info, err := s.manager.GetDeviceInfo(req.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device info: %w", err)
	}
	return info, nil
}

func (s *RPCServer) handleInitializeDevice(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		DeviceID   string `json:"device_id"`
		Passphrase string `json:"passphrase"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if err := s.manager.InitializeDevice(req.DeviceID, req.Passphrase); err != nil {
		return nil, fmt.Errorf("failed to initialize device: %w", err)
	}
	return "initialized", nil
}

func (s *RPCServer) handleUpdateFirmware(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		DeviceID     string `json:"device_id"`
		FirmwareData []byte `json:"firmware_data"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if err := s.manager.UpdateDeviceFirmware(req.DeviceID, req.FirmwareData); err != nil {
		return nil, fmt.Errorf("failed to update firmware: %w", err)
	}
	return "firmware update started", nil
}

func (s *RPCServer) handleSetPin(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		DeviceID string `json:"device_id"`
		Pin      string `json:"pin"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if err := s.manager.SetDevicePin(req.DeviceID, req.Pin); err != nil {
		return nil, fmt.Errorf("failed to set PIN: %w", err)
	}
	return "PIN set", nil
}

func (s *RPCServer) handleChangePin(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		DeviceID string `json:"device_id"`
		OldPin   string `json:"old_pin"`
		NewPin   string `json:"new_pin"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if err := s.manager.ChangeDevicePin(req.DeviceID, req.OldPin, req.NewPin); err != nil {
		return nil, fmt.Errorf("failed to change PIN: %w", err)
	}
	return "PIN changed", nil
}

func (s *RPCServer) handleUnlockDevice(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		DeviceID string `json:"device_id"`
		Pin      string `json:"pin"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if err := s.manager.UnlockDevice(req.DeviceID, req.Pin); err != nil {
		return nil, fmt.Errorf("failed to unlock device: %w", err)
	}
	return "unlocked", nil
}

func (s *RPCServer) handleLockDevice(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		DeviceID string `json:"device_id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if err := s.manager.LockDevice(req.DeviceID); err != nil {
		return nil, fmt.Errorf("failed to lock device: %w", err)
	}
	return "locked", nil
}

func (s *RPCServer) handleGetFirmwareVersion(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		DeviceID string `json:"device_id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	version, err := s.manager.GetDeviceFirmwareVersion(req.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get firmware version: %w", err)
	}
	return version, nil
}

func (s *RPCServer) handleGetFirmwareUpdateStatus(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		DeviceID string `json:"device_id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	status, err := s.manager.GetDeviceFirmwareUpdateStatus(req.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get firmware update status: %w", err)
	}
	return status, nil
}

func (s *RPCServer) handleGetBatteryLevel(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		DeviceID string `json:"device_id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	level, err := s.manager.GetDeviceBatteryLevel(req.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get battery level: %w", err)
	}
	return level, nil
}

func (s *RPCServer) handleSetPassphrase(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		DeviceID   string `json:"device_id"`
		Passphrase string `json:"passphrase"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if err := s.manager.SetDevicePassphrase(req.DeviceID, req.Passphrase); err != nil {
		return nil, fmt.Errorf("failed to set passphrase: %w", err)
	}
	return "passphrase set", nil
}

func (s *RPCServer) handleRemovePassphrase(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		DeviceID string `json:"device_id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if err := s.manager.RemoveDevicePassphrase(req.DeviceID); err != nil {
		return nil, fmt.Errorf("failed to remove passphrase: %w", err)
	}
	return "passphrase removed", nil
}

func (s *RPCServer) handleGetBackupStatus(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		DeviceID string `json:"device_id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	status, err := s.manager.GetDeviceBackupStatus(req.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup status: %w", err)
	}
	return status, nil
}

func (s *RPCServer) handleCreateBackup(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		DeviceID string `json:"device_id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if err := s.manager.CreateDeviceBackup(req.DeviceID); err != nil {
		return nil, fmt.Errorf("failed to create backup: %w", err)
	}
	return "backup created", nil
}

func (s *RPCServer) handleRestoreBackup(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		DeviceID   string `json:"device_id"`
		BackupData []byte `json:"backup_data"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if err := s.manager.RestoreDeviceBackup(req.DeviceID, req.BackupData); err != nil {
		return nil, fmt.Errorf("failed to restore backup: %w", err)
	}
	return "backup restored", nil
}

func (s *RPCServer) handleGetConnectedDevices(ctx context.Context, params json.RawMessage) (interface{}, error) {
	devices, err := s.manager.GetConnectedDevices()
	if err != nil {
		return nil, fmt.Errorf("failed to get connected devices: %w", err)
	}
	return devices, nil
}

func (s *RPCServer) handleGetDeviceStatus(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		DeviceID string `json:"device_id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	status, err := s.manager.GetDeviceStatus(req.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device status: %w", err)
	}
	return status, nil
}

func (s *RPCServer) handleGetDeviceCapabilities(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		DeviceID string `json:"device_id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	capabilities, err := s.manager.GetDeviceCapabilities(req.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device capabilities: %w", err)
	}
	return capabilities, nil
}

func (s *RPCServer) handleGetDeviceMetrics(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		DeviceID string `json:"device_id"`
		Start    int64  `json:"start"`
		End      int64  `json:"end"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	metrics, err := s.manager.GetDeviceMetrics(req.DeviceID, req.Start, req.End)
	if err != nil {
		return nil, fmt.Errorf("failed to get device metrics: %w", err)
	}
	return metrics, nil
}

func (s *RPCServer) handleGetDeviceLogs(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		DeviceID string `json:"device_id"`
		Start    int64  `json:"start"`
		End      int64  `json:"end"`
		Level    string `json:"level"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	logs, err := s.manager.GetDeviceLogs(req.DeviceID, req.Start, req.End, req.Level)
	if err != nil {
		return nil, fmt.Errorf("failed to get device logs: %w", err)
	}
	return logs, nil
}

func (s *RPCServer) handleGetDeviceEvents(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		DeviceID string `json:"device_id"`
		Start    int64  `json:"start"`
		End      int64  `json:"end"`
		Type     string `json:"type"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	events, err := s.manager.GetDeviceEvents(req.DeviceID, req.Start, req.End, req.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to get device events: %w", err)
	}
	return events, nil
}

func (s *RPCServer) handleExportPublicKey(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		DeviceID string `json:"device_id"`
		Path     string `json:"path"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	key, err := s.manager.ExportPublicKey(req.DeviceID, req.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to export public key: %w", err)
	}
	return key, nil
}

func (s *RPCServer) handleVerifyMessage(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		DeviceID  string `json:"device_id"`
		Message   string `json:"message"`
		Signature string `json:"signature"`
		Address   string `json:"address"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	valid, err := s.manager.VerifyMessage(req.DeviceID, req.Message, req.Signature, req.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to verify message: %w", err)
	}
	return valid, nil
}

func (s *RPCServer) handleSignMessage(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		DeviceID string `json:"device_id"`
		Message  string `json:"message"`
		Path     string `json:"path"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	signature, err := s.manager.SignMessage(req.DeviceID, req.Message, req.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to sign message: %w", err)
	}
	return signature, nil
}

func (s *RPCServer) handleGetXPub(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		DeviceID string `json:"device_id"`
		Path     string `json:"path"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	xpub, err := s.manager.GetXPub(req.DeviceID, req.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to get xpub: %w", err)
	}
	return xpub, nil
}

func (s *RPCServer) handleGetAddress(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		DeviceID string `json:"device_id"`
		Path     string `json:"path"`
		Type     string `json:"type"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	address, err := s.manager.GetAddress(req.DeviceID, req.Path, req.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to get address: %w", err)
	}
	return address, nil
}

func (s *RPCServer) handleGetAddressInfo(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		DeviceID string `json:"device_id"`
		Address  string `json:"address"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	info, err := s.manager.GetAddressInfo(req.DeviceID, req.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to get address info: %w", err)
	}
	return info, nil
}
