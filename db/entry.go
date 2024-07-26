package db

import (
	"strings"
	"velvet/perm"
	"velvet/session"

	"github.com/df-mc/dragonfly/server/player"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

// Entry is a player database entry.
type Entry struct {
	// XUID is the xuid of the player.
	XUID string `bson:"xuid"`
	// DisplayName is the display name of the player.
	DisplayName string `bson:"display_name"`
	// Name is the lowercase name of the player.
	Name string `bson:"name"`
	// DeviceID is the device id of the player.
	DeviceID string `bson:"device_id"`
	// Rank is the rank of the player.
	Rank string `bson:"rank"`
	// Kills are the amount of players the player has killed.
	Kills uint32 `bson:"kills"`
	// Deaths are the amount of times the player has died.
	Deaths uint32 `bson:"deaths"`
	// Punishments contains all the punishments of the player.
	Punishments Punishments `bson:"punishments"`
}

// Punishments contains the punishments of a player.
type Punishments struct {
	Ban  Punishment `bson:"ban"`
	Mute Punishment `bson:"mute"`
}

// Register registers a player into the database.
func Register(xuid, displayName, deviceID string) {
	return
	_, ok := sess.Collection("players").Find(nil, bson.D{{"_id", xuid}}, options.Find())
	if ok == nil {
		_, err := sess.Collection("players").UpdateOne(nil, bson.D{{"_id", xuid}}, map[string]string{"display_name": displayName, "name": strings.ToLower(displayName), "device_id": deviceID}, options.Update())
		if err != nil {
			panic(err)
		}
	} else {
		_, _ = sess.Collection("players").InsertOne(nil, bson.D{{"_id", xuid}, {"XUID", xuid}, {"DisplayName", displayName}, {"Name", strings.ToLower(displayName)}, {"DeviceID", deviceID}}, options.InsertOne())
	}
}

// LoadSession loads a user session from the database.
func LoadSession(p *player.Player) (*session.Session, error) {
	return session.New(p, perm.GetRank("Owner"), 0, 0, "deviceid"), nil
	entry, err := findPlayer(p.Name())
	if err != nil {
		return nil, err
	}

	return session.New(p,
		perm.GetRank(entry.Rank),
		entry.Kills,
		entry.Deaths,
		entry.DeviceID,
	), nil
}

// SaveSession saves a user session to the database.
func SaveSession(session *session.Session) error {
	return SaveOfflinePlayer(&Entry{
		XUID:        session.Player.XUID(),
		DisplayName: session.Player.Name(),
		Name:        strings.ToLower(session.Player.Name()),
		DeviceID:    session.DeviceID(),
		Rank:        session.RankName(),
		Kills:       session.Kills(),
		Deaths:      session.Deaths(),
		Punishments: Punishments{}, // todo
	})
}

// Registered returns whether a player is registered
func Registered(id string) bool {
	ok, _ := findPlayer(id)
	return ok != nil
}

// LoadOfflinePlayer returns an offline player entry for the given ign, if the player does not exist, an error will be returned.
func LoadOfflinePlayer(ign string) (*Entry, error) {
	return &Entry{
		DisplayName: ign,
		Name: strings.ToLower(ign),
	}, nil
	return findPlayer(ign)
}

// LoadOfflinePlayers returns all player entries that match the given conditions.
func LoadOfflinePlayers(cond interface{}) ([]*Entry, error) {
	return []*Entry{}, nil/*
	var data *bson.D
	cursor, err := sess.Collection("players").Find(nil, cond, options.Find())
	cursor.All(context.Background(), &data)
	var finmap []*Entry

	for _, v := range data.Map() {
		finmap = append(finmap, &Entry{
			XUID: v.XUID,
			DisplayName: v.DisplayName,
			Name: v.Name,
			DeviceID: v.DeviceID,
			Rank: v.Rank,
			Kills: v.Kills,
			Deaths: v.Deaths,
			Punishments: v.Punishments,
		})
	}
	return finmap, err*/
}

// SaveOfflinePlayer saves the entry of an offline player.
func SaveOfflinePlayer(entry *Entry) error {
	return nil
	players := sess.Collection("players")
	doc, e1 := EntrytoDoc(entry)
	if e1 != nil {
		return e1
	}
	_, ok := players.Find(nil, bson.D{{"_id", entry.XUID}})
	if ok == nil {
		_, err := sess.Collection("players").UpdateOne(nil, doc, options.Update())
		if err != nil {
			return err
		}
	} else {
		var err error
		_, err = sess.Collection("players").InsertOne(nil, doc, options.InsertOne())
		if err != nil {
			return err
		}
	}
	return nil
}

// GetAlias will return all the names that have the same deviceID as the given ign.
// Zero values will be returned if the player has never joined before.
func GetAlias(ign string) (deviceID string, names []string) {
	return "device", []string{ign}
	p, err := LoadOfflinePlayer(ign)
	if err != nil {
		return
	}
	deviceID = p.DeviceID
	entries, _ := LoadOfflinePlayers(bson.D{{"DeviceID", deviceID}})
	for _, e := range entries {
		name := "§e" + e.DisplayName
		if !e.Punishments.Ban.Expired() {
			name += " §l§cBANNED§r"
		}
		names = append(names, name)
	}
	return
}

// IsStaff will return whether a player has a staff rank.
func IsStaff(id string) bool {
	return true
	p, err := LoadOfflinePlayer(id)
	return err == nil && perm.StaffRanks.Contains(p.Rank)
}

// findPlayer is used internally to fetch a player entry from the database.
func findPlayer(id string) (entry *Entry, err error) {
	return
	byxuid, err := sess.Collection("players").Find(nil, bson.D{{"_id", id}}, options.Find())
	if err == nil {
		entry, err = CursorToEntry(byxuid)
		return
	} else {
		byname, err := sess.Collection("players").Find(nil, bson.D{{"Name", id}}, options.Find())
		if err == nil {
			entry, err = CursorToEntry(byname)
		}
	}
	return
}

func EntrytoDoc(v interface{}) (doc *bson.D, err error) {
    data, err := bson.Marshal(v)
    if err != nil {
        return
    }

    err = bson.Unmarshal(data, &doc)
    return
}

func CursorToEntry(cursor *mongo.Cursor) (entry *Entry, err error) {
	return/*
	var all *bson.D
	err = cursor.All(context.Background(), &all)
	if err != nil {
		return
	}

	mapp := all.Map()
	var fk string
	for k, _ := range mapp {
		fk = k
		break
	}

	first := mapp[fk]
	log.Print(first)
	entry = &Entry{
		XUID: first.XUID,
		DisplayName: first.DisplayName,
		Name: first.Name,
		DeviceID: first.DeviceID,
		Rank: first.Rank,
		Kills: first.Kills,
		Deaths: first.Deaths,
		Punishments: first.Punishments,
	}

	return*/
}
