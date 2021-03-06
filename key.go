package amber

import "github.com/golang/glog"

var (
	_ KeyNet = (*BaseKey)(nil)
)

type Key interface {
	handle(req *Req, cmd string, args ...interface{})
}

type KeyNet interface {
	Master() string
	Head() uint32
	AddDiff(diff *Diff)
}

type BaseKey struct {
	master  string
	head    uint32
	commits map[uint32]*Commit
	diffs   map[uint32]*Diff
	deleted bool
}

func (bk *BaseKey) Master() string {
	return bk.master
}

func (bk *BaseKey) SetMaster(master string) {
	bk.master = master
}

func (bk *BaseKey) Head() uint32 {
	return bk.head
}

func (bk *BaseKey) AddDiff(diff *Diff) {
	bk.diffs[diff.commitId] = diff
}

func NewBaseKey(master string) *BaseKey {
	return &BaseKey{
		master:  master,
		commits: make(map[uint32]*Commit),
		diffs:   make(map[uint32]*Diff),
	}
}

func KeyHandler(db *DB, req *Req) {
	mode := req.options[0].(int)
	var newKeyFunc func(master string) Key
	if mode == KeyCmdModeUpsert {
		newKeyFunc = req.options[1].(func(master string) Key)
	} else {
		newKeyFunc = nil
	}

	keyName := req.options[2].(string)
	level := req.options[3].(int)
	cmd := req.options[4].(string)
	args := req.options[5:]

	glog.V(2).Infof("mode: %d, level: %d, cmd: %s, keyName: %s, args: %v", mode, level, cmd, keyName, args)

	key, err := db.load(keyName, level)
	if err != nil && (err.Error() != ErrUndefinedKey || mode != KeyCmdModeUpsert) {
		req.res <- NewEmptyRes(err)
	} else if err != nil {
		key = newKeyFunc(req.master)
		go db.add(keyName, key, level)
	}

	key.handle(req, cmd, args...)
}
