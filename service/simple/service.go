package simple

import (
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/micromdm/checkin"
	"github.com/micromdm/mdm"
	nsq "github.com/nsqio/go-nsq"
	"golang.org/x/net/context"
)

// CheckinBucket is the *bolt.DB bucket where checkins are archived.
const CheckinBucket = "mdm.Checkin.ARCHIVE"

// NSQ Topics where MDM Checkin events are published to
const (
	AuthenticateTopic = "mdm.Authenticate"
	TokenUpdateTopic  = "mdm.TokenUpdate"
	CheckoutTopic     = "mdm.CheckOut"
)

// The publisher interface is satisfied by an NSQ producer.
// Only used in tests.
type publisher interface {
	Publish(string, []byte) error
}

// archiveFunc is the function signature for archiving events in BoltDB.
// CheckinService.archive is used outside of tests.
type archiveFunc func(int64, []byte) error

// CheckinService implements the MDM Check-in protocol and responds to Check-in
// requests and publishes them to an NSQ topic.
// The CheckinService also archives all request to a BoltDB bucket.
type CheckinService struct {
	db *bolt.DB
	publisher

	archiveFn archiveFunc
}

// NewService creates a CheckinService.
func NewService(db *bolt.DB, producer *nsq.Producer) (*CheckinService, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(CheckinBucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	svc := &CheckinService{db: db, publisher: producer}
	svc.archiveFn = svc.archive
	return svc, nil
}

func (svc *CheckinService) Authenticate(ctx context.Context, cmd mdm.CheckinCommand) error {
	if cmd.MessageType != "Authenticate" {
		return fmt.Errorf("expected Authenticate, got %s MessageType", cmd.MessageType)
	}
	return svc.archiveAndPublish(AuthenticateTopic, cmd)
}

func (svc *CheckinService) TokenUpdate(ctx context.Context, cmd mdm.CheckinCommand) error {
	if cmd.MessageType != "TokenUpdate" {
		return fmt.Errorf("expected TokenUpdate, got %s MessageType", cmd.MessageType)
	}
	return svc.archiveAndPublish(TokenUpdateTopic, cmd)
}

func (svc *CheckinService) CheckOut(ctx context.Context, cmd mdm.CheckinCommand) error {
	if cmd.MessageType != "CheckOut" {
		return fmt.Errorf("expected CheckOut, but got %s MessageType", cmd.MessageType)
	}
	return svc.archiveAndPublish(CheckoutTopic, cmd)
}

func (svc *CheckinService) archiveAndPublish(topic string, cmd mdm.CheckinCommand) error {
	event := checkin.NewEvent(cmd)
	msg, err := checkin.MarshalEvent(event)
	if err != nil {
		return err
	}
	if err := svc.archiveFn(event.Time.UnixNano(), msg); err != nil {
		return err
	}
	if err := svc.Publish(topic, msg); err != nil {
		return err
	}
	return nil
}

// archive events to BoltDB bucket using timestamp as key to preserve order.
func (svc *CheckinService) archive(nano int64, msg []byte) error {
	tx, err := svc.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bkt := tx.Bucket([]byte(CheckinBucket))
	if bkt == nil {
		return fmt.Errorf("bucket %q not found!", CheckinBucket)
	}
	key := []byte(fmt.Sprintf("%d", nano))
	if err := bkt.Put(key, msg); err != nil {
		return err
	}
	return tx.Commit()
}
