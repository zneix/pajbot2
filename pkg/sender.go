package pkg

type Sender interface {
	Say(Channel, string)
	Mention(Channel, User, string)
	Timeout(Channel, User, int, string)
	GetPoints(Channel, User) uint64

	// give or remove points from user in channel
	BulkEdit(string, []string, int32)

	AddPoints(Channel, string, uint64) (bool, uint64)
	RemovePoints(Channel, string, uint64) (bool, uint64)
	ForceRemovePoints(Channel, string, uint64) uint64

	GetUserStore() UserStore
}