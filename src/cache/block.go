package cache

import (
	"encoding/hex"
	"os"
	"../fs"
	"../config"
)

type Block struct {
	Id     int64
	Hash   string // Ciphertext hash
	Remote bool   // Available on cloud server
	Tail   bool   // Is this an unpacked short block?
	Dead   bool   // No longer used, should be deleted
}

func (db *DB) connectBlocks() {
	tab := db.dbm.AddTableWithName(Block{}, "blocks")
	tab.SetKeys(true, "Id")
	tab.ColMap("Hash").SetUnique(true).SetNotNull(true)
}

func (db *DB) FindBlock(hash []byte) *Block {
	var bs []Block

	_, err := db.dbm.Select(
		&bs, 
		"select * from blocks where Hash = ? limit 1",
		hex.EncodeToString(hash))
	fs.CheckError(err)

	if len(bs) == 0 {
		return nil
	} else {
		return &bs[0]
	}
}
func (bb *Block) GetHash() []byte {
	hash, err := hex.DecodeString(bb.Hash)
	fs.CheckError(err)
	return hash
}

func (bb *Block) SetHash(hash []byte) {
	bb.Hash = hex.EncodeToString(hash)
}

func (bb *Block) Cached() bool {
	hash := bb.GetHash()
	bpath := config.BlockPath(hash)

	_, err := os.Lstat(bpath)
	if os.IsNotExist(err) {
		return false
	}
	fs.CheckError(err)
	
	return true
}
