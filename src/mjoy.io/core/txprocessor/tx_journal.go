package txprocessor

import (
	"errors"
	"io"
	"os"
	"mjoy.io/common/types"
	"github.com/tinylib/msgp/msgp"
	"mjoy.io/core/transaction"
)

// errNoActiveJournal is returned if a transaction is attempted to be inserted
// into the journal, but no such file is currently open.
var errNoActiveJournal = errors.New("no active journal")

// devNull is a WriteCloser that just discards anything written into it. Its
// goal is to allow the transaction journal to write into a fake journal when
// loading transactions on startup without printing warnings due to no file
// being readt for write.
type devNull struct{}

func (*devNull) Write(p []byte) (n int, err error) { return len(p), nil }
func (*devNull) Close() error                      { return nil }

// txJournal is a rotating log of transactions with the aim of storing locally
// created transactions to allow non-executed ones to survive node restarts.
type txJournal struct {
	path   string         // Filesystem path to store the transactions at
	writer io.WriteCloser // Output stream to write new transactions into
}

// newTxJournal creates a new transaction journal to
func newTxJournal(path string) *txJournal {
	return &txJournal{
		path: path,
	}
}

// load parses a transaction journal dump from disk, loading its contents into
// the specified pool.
func (journal *txJournal) load(add func(*transaction.Transaction) error) error {
	// Skip the parsing if the journal file doens't exist at all
	if _, err := os.Stat(journal.path); os.IsNotExist(err) {
		return nil
	}
	// Open the journal for loading any past transactions
	input, err := os.Open(journal.path)
	if err != nil {
		return err
	}
	defer input.Close()

	// Temporarily discard any journal additions (don't double add on load)
	journal.writer = new(devNull)
	defer func() { journal.writer = nil }()

	// Inject all transactions from the journal into the pool


	total, dropped := 0, 0

	var failure error
	for {
		// Parse the next transaction and terminate on error
		tx := new(transaction.Transaction)

		if err = msgp.Decode(input,tx); err != nil {
			if err != io.EOF {
				failure = err
			}
			break
		}

		// Import the transaction and bump the appropriate progress counters
		total++
		tx.PrintDataInfo()
		if err = add(tx); err != nil {
			logger.Debug("Failed to add journaled transaction", "err", err)
			dropped++
			continue
		}
	}
	logger.Info("Loaded local transaction journal", "transactions", total, "dropped", dropped)

	return failure
}

// insert adds the specified transaction to the local disk journal.
func (journal *txJournal) insert(tx *transaction.Transaction) error {
	if journal.writer == nil {
		return errNoActiveJournal
	}

	if err := msgp.Encode(journal.writer,tx); err != nil {
		return err
	}
	return nil
}

// rotate regenerates the transaction journal based on the current contents of
// the transaction pool.
func (journal *txJournal) rotate(all map[types.Address]transaction.Transactions) error {
	// Close the current journal (if any is open)
	if journal.writer != nil {
		if err := journal.writer.Close(); err != nil {
			return err
		}
		journal.writer = nil
	}
	// Generate a new journal with the contents of the current pool
	replacement, err := os.OpenFile(journal.path+".new", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	journaled := 0
	for _, txs := range all {
		for _, tx := range txs {


			if err = msgp.Encode(replacement,tx); err != nil {
				replacement.Close()
				return err
			}
		}
		journaled += len(txs)
	}
	replacement.Close()

	// Replace the live journal with the newly generated one
	if err = os.Rename(journal.path+".new", journal.path); err != nil {
		return err
	}
	sink, err := os.OpenFile(journal.path, os.O_WRONLY|os.O_APPEND, 0755)
	if err != nil {
		return err
	}
	journal.writer = sink
	logger.Info("Regenerated local transaction journal", "transactions", journaled, "accounts", len(all))

	return nil
}

// close flushes the transaction journal contents to disk and closes the file.
func (journal *txJournal) close() error {
	var err error

	if journal.writer != nil {
		err = journal.writer.Close()
		journal.writer = nil
	}
	return err
}