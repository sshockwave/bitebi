package message

// The “getheaders” message is nearly identical to the “getblocks” message,
// with one minor difference:
// the inv reply to the “getblocks” message will include no more than 500 block header hashes;
// the headers reply to the “getheaders” message will include as many as 2,000 block headers.

// TODO: This is not yet implemented since it can be replaced by getblocks
