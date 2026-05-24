package constant

// Op represents a CDC operation type following the Debezium format.
type Op string

const (
	OpCreate   Op = "c" // Insert
	OpUpdate   Op = "u" // Update
	OpDelete   Op = "d" // Delete
	OpSnapshot Op = "r" // Read (snapshot)
)

func (o Op) String() string {
	return string(o)
}
