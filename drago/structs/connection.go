package structs

import (
	"errors"
	"sort"
	"time"

	"github.com/seashell/drago/pkg/uuid"
)

// Connection :
type Connection struct {
	ID        string
	NetworkID string

	// PeerSettings contains the ID and the configurations to be applied
	// to each of the connected interfaces.
	PeerSettings []*PeerSettings

	// If the connection is going from a NAT-ed peer to a public peer,
	// the node behind the NAT must regularly send an outgoing ping to
	// keep the bidirectional connection alive in the NAT router's
	// connection table.
	PersistentKeepalive *int

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewConnection() *Connection {

	c := &Connection{}

	c.ID = uuid.Generate()
	c.CreatedAt = time.Now()

	return c
}

// Validate :
func (c *Connection) Validate() error {

	connectedInterfaceIDs := c.ConnectedInterfaceIDs()

	if len(connectedInterfaceIDs) != 2 {
		return errors.New("a connection must specify exactly two interfaces")
	}
	if connectedInterfaceIDs[0] == connectedInterfaceIDs[1] {
		return errors.New("can't connect an interface to itself")
	}

	return nil
}

// ConnectedInterfaceIDs :
func (c *Connection) ConnectedInterfaceIDs() []string {
	ids := []string{}
	for _, peer := range c.PeerSettings {
		ids = append(ids, peer.InterfaceID)
	}
	sort.Strings(ids)
	return ids
}

// ConnectedNodeIDs :
func (c *Connection) ConnectedNodeIDs() []string {
	ids := []string{}
	for _, peer := range c.PeerSettings {
		ids = append(ids, peer.NodeID)
	}
	sort.Strings(ids)
	return ids
}

// PeerSettingsByNodeID :
func (c *Connection) PeerSettingsByNodeID(s string) *PeerSettings {

	if c.PeerSettings[0].NodeID == s {
		return c.PeerSettings[0]
	} else if c.PeerSettings[1].NodeID == s {
		return c.PeerSettings[1]
	}

	return nil
}

// PeerSettingsByInterfaceID :
func (c *Connection) PeerSettingsByInterfaceID(s string) *PeerSettings {

	if c.PeerSettings[0].InterfaceID == s {
		return c.PeerSettings[0]
	} else if c.PeerSettings[1].InterfaceID == s {
		return c.PeerSettings[1]
	}

	return nil
}

// OtherPeerSettingsByInterfaceID : given the ID of one of the connected interfaces,
// returns the settings for the peer/interface at the other end of the connection.
func (c *Connection) OtherPeerSettingsByInterfaceID(s string) *PeerSettings {

	if c.PeerSettings[0].InterfaceID == s {
		return c.PeerSettings[1]
	} else if c.PeerSettings[1].InterfaceID == s {
		return c.PeerSettings[0]
	}

	return nil
}

// ConnectsInterfaces : checks whether a Connection connects two
// interfaces whose indices are passed as arguments.
func (c *Connection) ConnectsInterfaces(a, b string) bool {
	if c.ConnectsInterface(a) && c.ConnectsInterface(b) {
		return true
	}
	return false
}

// ConnectsInterface : checks whether a connection connects
// an interface whose index is passed as argument.
func (c *Connection) ConnectsInterface(s string) bool {

	if c.PeerSettings[0].InterfaceID == s || c.PeerSettings[1].InterfaceID == s {
		return true
	}
	return false
}

func (c *Connection) InitializePeerSettings() error {

	for _, id := range c.ConnectedInterfaceIDs() {

		if c.PeerSettingsByInterfaceID(id) == nil {
			c.PeerSettings = append(c.PeerSettings, &PeerSettings{
				InterfaceID: id,
				RoutingRules: &RoutingRules{
					AllowedIPs: []string{},
				},
			})
		}

		// Initialize RoutingRules, if necessary
		peer := c.PeerSettingsByInterfaceID(id)
		if peer.RoutingRules == nil {
			peer.RoutingRules = &RoutingRules{AllowedIPs: []string{}}
		}
	}

	return nil

}

// Merge :
func (c *Connection) Merge(in *Connection) *Connection {

	result := *c

	if in.PeerSettings != nil {
		if result.PeerSettings == nil {
			result.PeerSettings = in.PeerSettings
		} else {
			for _, peer := range in.PeerSettings {
				if result.PeerSettings[0].InterfaceID == peer.InterfaceID {
					result.PeerSettings[0] = result.PeerSettings[0].Merge(peer)
				} else if result.PeerSettings[1].InterfaceID == peer.InterfaceID {
					result.PeerSettings[1] = result.PeerSettings[1].Merge(peer)
				}
			}
		}
	}

	if in.PersistentKeepalive != nil {
		result.PersistentKeepalive = in.PersistentKeepalive
	}

	return &result
}

func (c *Connection) AllowIPBidirectional(ip string) error {
	for _, peer := range c.PeerSettings {
		peer.RoutingRules.AllowedIPs = append(peer.RoutingRules.AllowedIPs, ip)
	}
	return nil
}

// Stub :
func (c *Connection) Stub() *ConnectionListStub {

	peers := []string{}
	for _, peer := range c.PeerSettings {
		peers = append(peers, peer.InterfaceID)
	}

	return &ConnectionListStub{
		ID:                  c.ID,
		NetworkID:           c.NetworkID,
		Peers:               peers,
		PeerSettings:        c.PeerSettings,
		PersistentKeepalive: c.PersistentKeepalive,
		BytesTransferred:    0,
		CreatedAt:           c.CreatedAt,
		UpdatedAt:           c.UpdatedAt,
	}
}

// ConnectionListStub :
type ConnectionListStub struct {
	ID                  string
	NetworkID           string
	NodeIDs             []string
	Peers               []string
	PeerSettings        []*PeerSettings
	PersistentKeepalive *int
	BytesTransferred    uint64
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// PeerSettings :
type PeerSettings struct {
	NodeID       string
	InterfaceID  string
	RoutingRules *RoutingRules
}

// Merge :
func (r *PeerSettings) Merge(in *PeerSettings) *PeerSettings {
	result := *r
	if in.NodeID != "" {
		result.NodeID = in.NodeID
	}
	if in.InterfaceID != "" {
		result.InterfaceID = in.InterfaceID
	}
	if in.RoutingRules != nil {
		result.RoutingRules = r.RoutingRules.Merge(in.RoutingRules)
	}
	return &result
}

// RoutingRules :
type RoutingRules struct {
	// AllowedIPs defines the IP ranges for which traffic will be routed/accepted.
	// Example: If AllowedIPs = [192.0.2.3/32, 192.168.1.1/24], the node
	// will accept traffic for itself (192.0.2.3/32), and for all nodes in the
	// local network (192.168.1.1/24).
	AllowedIPs []string
}

// Merge :
func (r *RoutingRules) Merge(in *RoutingRules) *RoutingRules {
	result := *r
	if in.AllowedIPs != nil {
		result.AllowedIPs = in.AllowedIPs
	}
	return &result
}

// ConnectionSpecificRequest :
type ConnectionSpecificRequest struct {
	ConnectionID string

	QueryOptions
}

// SingleConnectionResponse :
type SingleConnectionResponse struct {
	Connection *Connection

	Response
}

// ConnectionUpsertRequest :
type ConnectionUpsertRequest struct {
	Connection *Connection

	WriteRequest
}

// ConnectionDeleteRequest :
type ConnectionDeleteRequest struct {
	ConnectionIDs []string

	WriteRequest
}

// ConnectionListRequest :
type ConnectionListRequest struct {
	InterfaceID string
	NodeID      string
	NetworkID   string

	QueryOptions
}

// ConnectionListResponse :
type ConnectionListResponse struct {
	Items []*ConnectionListStub

	Response
}
