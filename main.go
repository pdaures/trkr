package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

func main() {
	var port int
	var mongoAddr string
	var mongoDB string
	var mongoCollection string

	flag.IntVar(&port, "port", 8080, "HTTP port")
	flag.StringVar(&mongoAddr, "mongoAddr", "", "MongoDB addr")
	flag.StringVar(&mongoDB, "mongoDatabase", "", "MongoDB database")
	flag.StringVar(&mongoCollection, "mongoCollection", "", "MongoDB collection")

	flag.Parse()

	mustBe("mongoAddr", mongoAddr)
	mustBe("mongoDatabase", mongoDB)
	mustBe("mongoCollection", mongoCollection)

	mux := http.NewServeMux()
	storer := &mongoStorer{
		addr:       mongoAddr,
		db:         mongoDB,
		collection: mongoCollection,
	}
	if err := storer.test(); err != nil {
		fmt.Printf("Cannot connect to MongoDB, %v\n", err)
		os.Exit(1)
	}

	mux.HandleFunc("/trck", track(storer))

	fmt.Printf("starting trck listening on http:<addr>:%d/trck/:userid\n", port)
	err := http.ListenAndServe(":"+strconv.Itoa(port), mux)
	if err != nil {
		fmt.Printf("Cannot start HTTP server, %v\n", err)
		os.Exit(1)
	}
}

func mustBe(name, val string) {
	if val == "" {
		fmt.Printf("%s must be set\n", name)
		os.Exit(1)
	}
}

func track(storer storer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := extractUser(r.URL.EscapedPath())
		if err != nil {
			fmt.Printf("bad request, %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		ip, err := extractIP(r)
		if err != nil {
			fmt.Printf("cannot extract the IP from the request, %v\n", err)
		}
		err = storer.store(Record{
			ID:        bson.NewObjectId(),
			Timestamp: time.Now(),
			URL:       r.URL.String(),
			User:      user,
			UserAgent: r.UserAgent(),
			IP:        ip,
		})
		if err != nil {
			fmt.Printf("cannot store the tracking request, %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func extractIP(r *http.Request) (string, error) {
	// if the original user is behind a proxy, then the original IP is in the header X-Forwarded-For
	if source := r.Header.Get(http.CanonicalHeaderKey("X-Forwarded-For")); source != "" {
		return source, nil
	}
	// Otherwise, the source IP is in request.RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	return ip, err
}

func extractUser(path string) (string, error) {
	paths := strings.Split(path, "/")
	if len(paths) != 3 {
		return "", fmt.Errorf("invalid path %s", path)
	}
	user := paths[2]
	if len(user) == 0 {
		return "", fmt.Errorf("invalid path %s", path)
	}
	return user, nil
}

// Record holds the tracking information of 1 visit
type Record struct {
	ID        bson.ObjectId `bson:"_id,omitempty"`
	Timestamp time.Time
	URL       string
	User      string
	IP        string
	UserAgent string
}

type storer interface {
	store(record Record) error
}

type mongoStorer struct {
	addr       string
	db         string
	collection string
}

func (m *mongoStorer) test() error {
	session, err := mgo.Dial(m.addr)
	if err != nil {
		return err
	}
	defer session.Close()
	return session.Ping()
}

func (m *mongoStorer) store(record Record) error {
	session, err := mgo.Dial(m.addr)
	if err != nil {
		return err
	}
	// For storing tracking events, the exact ordering doesn't matter
	session.SetMode(mgo.Eventual, false)
	defer session.Close()
	c := session.DB(m.db).C(m.collection)
	return c.Insert(record)
}
