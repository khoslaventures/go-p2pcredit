# go-p2pcredit - a peer-to-peer credit messaging system

Clone this repo and run:
```
go build
```

To run locally only (publishes `localhost` as the IP address in `peering_info` to Fakechain):
```
./messages --port PORT_NUMBER --local USERNAME STARTING_BALANCE
```

Omit the `--local` flag to allow connections beyond localhost:
```
./messages --port PORT_NUMBER USERNAME STARTING_BALANCE
```

Once launched, you will be prompted for a password. This is just the private
key for Fakechain. Right now, it's just stored in memory because we don't
require persistence, and it's never asked for again.

**Available Commands**

```
Command options:
pay <peerID> <amount> - pays peerID the amount in a trustline
settle <peerID> <amount> - settles amount on Fakechain with peerID for trustline
propose <peerID> - proposes a trustline to peerID
balance - displays peerID and corresponding trustline balance
users - query Fakechain for user information
exit - settle as much debt as possible and exit
delete - deletes all users
```

## Potential Issues/Needs

- Since trustlines deal in credit, the host can pay (accumlate debt) to the point
  where their Fakechain balance wouldn't be able to settle all the debt. This is
  a potential issue if you wanted to exit and settle with everyone to find out
  you can't settle.
- Cleaner, more focused modularity. I'd spend more time on this and make it more
  friendly, but because no one is using this, I've quickly thrown things into
  various .go files with not as much thought as a production system.
- Better way to print `> ` to prompt.
- Better CLI handling/interface
- Error handling should be consistent (i.e. Fakechain's API calls), and tested
- Testing for all functionality
- Clear comments for each component
- Easy script for ngrok


## Motivation

I wrote this in Golang because I helped out a lot with Starlight, the Stellar
payment channels implementation. It's the closest thing I've written to this,
because client and server are in one package. However, Starlight is designed for
production, and so much thought was put into just how to design the state
machine that I had to condense what I remembered in that process down to a mere
3 days. This was tough, because I had to fight my urge to think in terms of
production code.

I wouldn't say I'm a master at Golang. I was going to write this
in Python, but the concurrency ergonomics are not nearly as powerful and
elegant. Golang offers goroutines and channels, and you can effectively write an application
without any explicit locking. Just use channels to synchronize. This was tough
to design. Starting off the scaffolding was probably the hardest part.

Definitely took me more than 12 hours because my VSCode debugger stopped working
(GOPATH issues, but vim reads it fine), and Go is a secondary language for me.

Overall, lesson is to just use Python or Node for coding challenges. Golang is
far better for writing production grade systems, because it encourages
minimalism, and that can be hard to initially build.


