package data

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgraph-io/badger"
	null "gopkg.in/guregu/null.v3"
	"gopkg.in/guregu/null.v3/zero"
)

// User represents a user in the system.  Users
// are associated with resources and roles within those applications/resources/services.
// They can be created/updated/deleted.  If they are deleted, eventually
// they will be removed from the system.  The admin user can only be disabled, not deleted
type User struct {
	Name        string      `json:"name"`
	Enabled     bool        `json:"enabled"`
	Description string      `json:"description"`
	SecretHash  string      `json:"secrethash"`
	Created     time.Time   `json:"created"`
	CreatedBy   string      `json:"created_by"`
	Updated     time.Time   `json:"updated"`
	UpdatedBy   string      `json:"updated_by"`
	Deleted     zero.Time   `json:"deleted"`
	DeletedBy   null.String `json:"deleted_by"`
}

// AddUser adds a user to the system
func (store Manager) AddUser(context User, user User, userPassword string) (User, error) {
	//	Our return item
	retval := User{}

	//	Open the systemDB
	opts := badger.DefaultOptions
	opts.Dir = store.SystemDBpath
	opts.ValueDir = store.SystemDBpath
	db, err := badger.Open(opts)
	if err != nil {
		return retval, fmt.Errorf("Problem opening the systemDB: %s", err)
	}
	defer db.Close()

	//	Update the created / updated fields:
	user.Created = time.Now()
	user.Updated = time.Now()
	user.CreatedBy = context.Name
	user.UpdatedBy = context.Name

	//	Serialize to JSON format
	encoded, err := json.Marshal(user)
	if err != nil {
		return retval, fmt.Errorf("Problem serializing the data: %s", err)
	}

	//	Save it to the database:
	err = db.Update(func(txn *badger.Txn) error {
		err := txn.Set(store.GetKey("User", user.Name), encoded)
		return err
	})

	//	If there was an error saving the data, report it:
	if err != nil {
		return retval, fmt.Errorf("Problem saving the data: %s", err)
	}

	//	Set our retval:
	retval = user

	//	Return our data:
	return retval, nil
}

// GetUser gets a user from the system
func (store Manager) GetUser(context User, userName string) (User, error) {
	//	Our return item
	retval := User{}

	//	Open the systemDB
	opts := badger.DefaultOptions
	opts.Dir = store.SystemDBpath
	opts.ValueDir = store.SystemDBpath
	db, err := badger.Open(opts)
	if err != nil {
		return retval, fmt.Errorf("Problem opening the systemDB: %s", err)
	}
	defer db.Close()

	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(store.GetKey("User", userName))
		if err != nil {
			return err
		}
		val, err := item.Value()
		if err != nil {
			return err
		}

		if len(val) > 0 {
			//	Unmarshal data into our item
			if err := json.Unmarshal(val, &retval); err != nil {
				return err
			}
		}

		return nil
	})

	//	If there was an error, report it:
	if err != nil {
		return retval, fmt.Errorf("Problem getting the data: %s", err)
	}

	//	Return our data:
	return retval, nil
}

// GetAllUsers gets all users in the system
func (store Manager) GetAllUsers(context User) ([]User, error) {
	//	Our return item
	retval := []User{}

	//	Open the systemDB
	opts := badger.DefaultOptions
	opts.Dir = store.SystemDBpath
	opts.ValueDir = store.SystemDBpath
	db, err := badger.Open(opts)
	if err != nil {
		return retval, fmt.Errorf("Problem opening the systemDB: %s", err)
	}
	defer db.Close()

	err = db.View(func(txn *badger.Txn) error {

		//	Get an iterator
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		//	Set our prefix
		prefix := store.GetKey("User")

		//	Iterate over our values:
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {

			//	Get the item
			item := it.Item()

			//	Get the item key
			// k := item.Key()

			//	Get the item value
			val, err := item.Value()
			if err != nil {
				return err
			}

			if len(val) > 0 {
				//	Create our item:
				item := User{}

				//	Unmarshal data into our item
				if err := json.Unmarshal(val, &item); err != nil {
					return err
				}

				//	Add to the array of returned users:
				retval = append(retval, item)
			}
		}
		return nil
	})

	//	If there was an error, report it:
	if err != nil {
		return retval, fmt.Errorf("Problem getting the list of items: %s", err)
	}

	//	Return our data:
	return retval, nil
}
