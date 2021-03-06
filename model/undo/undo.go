package undo

import (
	"bytes"
	"io"
	"github.com/copernet/copernicus/model/utxo"
	"github.com/copernet/copernicus/model/tx"
	"github.com/copernet/copernicus/util"
)

const MaxInputPerTx = tx.MaxTxInPerMessage

type DisconnectResult int

const (
	// DisconnectOk All good.
	DisconnectOk DisconnectResult = iota
	// DisconnectUnclean Rolled back, but UTXO set was inconsistent with block.
	DisconnectUnclean
	// DisconnectFailed Something else went wrong.
	DisconnectFailed
)



type TxUndo struct {
	undoCoins []*utxo.Coin
}


func (tu *TxUndo) SetUndoCoins(coins []*utxo.Coin){
	 tu.undoCoins = coins
}

func (tu *TxUndo) GetUndoCoins() []*utxo.Coin{
	return tu.undoCoins
}

func (tu *TxUndo) Serialize(w io.Writer) error {
	err := util.WriteVarInt(w, uint64(len(tu.undoCoins)))
	if err != nil {
		return err
	}
	for _, coin := range tu.undoCoins {
		err = coin.Serialize(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (tu *TxUndo)Unserialize(r io.Reader) error {

	 count, err := util.ReadVarInt(r)
	 if err != nil{
	 	return err
	 }
	if count > MaxInputPerTx {
		panic("Too many input undo records")
	}
	preouts := make([]*utxo.Coin, count,count)
	for i:=0; i<int(count); i++{
		coin := utxo.NewEmptyCoin()
		err := coin.Unserialize(r)

		if err != nil {
			return err
		}
		preouts[i] = coin
	}
	tu.undoCoins = preouts
	return nil
}



type BlockUndo struct {
	txundo []*TxUndo
}

func (bu *BlockUndo)GetTxundo()[]*TxUndo{
	return bu.txundo
}

func NewBlockUndo(count int) *BlockUndo {
	return &BlockUndo{
		txundo: make([]*TxUndo, 0, count),
	}
}

func (bu *BlockUndo) Serialize(w io.Writer) error {
	count := len(bu.txundo)
	util.WriteVarLenInt(w, uint64(count))
	for _, obj := range bu.txundo{
		err := obj.Serialize(w)
		return err
	}
	return nil

}

func (bu *BlockUndo) SerializeSize() int {
	buf := bytes.NewBuffer(nil)
	bu.Serialize(buf)
	return buf.Len()
	
}

func (bu *BlockUndo) Unserialize(r io.Reader) error {
	count, err := util.ReadVarLenInt(r)
	txundos := make([]*TxUndo, count, count)
	for i := 0; i<int(count); i++{
		obj := NewTxUndo()
		err = obj.Unserialize(r)
		if err != nil{
			return err
		}
		txundos[i] =  obj
	}
	bu.txundo = txundos
	return nil
}

func (bu *BlockUndo) SetTxUndo(txUndo []*TxUndo){
	bu.txundo = txUndo
}
func (bu *BlockUndo) AddTxUndo(txUndo *TxUndo){
	bu.txundo = append(bu.txundo,txUndo)
}



func NewTxUndo() *TxUndo {
	return &TxUndo{
		undoCoins: make([]*utxo.Coin, 0),
	}
}
