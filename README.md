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


