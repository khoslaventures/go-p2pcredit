
# Strata Labs Technical Challenge

## Logistics
This challenge is intended to be done at home and should take 3-6 hours. You 
should complete it in whatever programming language you feel most comfortable 
with and you may use any resources you like.

We will schedule a review session afterwards where we will ask you to discuss 
your solution and explain your design choices.

Upon completion, please send a zip of all the files used in your solution to 
```info@stratalabs.io```.

_Please do not share this challenge or your solution to it._

## Challenge
The year is 2009 and settling over a blockchain is slow and expensive.  With the
invention of decentralized financial ledgers, we now have a trustless way to
communicate value.  However, we still are not able to send micropayments because
of the cost of settling over blockchains.  You will be designing a decentralized
credit lending network that facilitates the communication of micropayments.

Participants in this network will use "trustlines" to maintain balances between
one another before settling on the underlying blockchain.  A trustline is a way
to keep track of debt between two parties.  The trustline balance will start at
0 and both parties independently track their views of the balance.  If Alice
sends 10 units to Bob, they would both update their trustline balance so that
Alice would see { Bob: -10 } and Bob would see { Alice: 10 }.  If Alice sends 10
more units to Bob, she would see a balance of { Bob: -20 } and he would see {
Alice: 20 }.  A negative amount indicates that a user owes money to the
counterparty, while a positive indicates that a user is owed money by the
counterparty.

Participants in this network have decided on a credit limit of 100 units.  That
means that anytime a trustline balance exceeds 100 units, that debt should be
settled on the underlying blockchain.  If Alice maintains balances of { Bob:
-60, Charlie: -95 } and wants to send Charlie 10 units, then a settlement would
be necessary because 105 exceeds the maximum trustline limit of 100.  Any time
the limit of 100 will be exceeded in a payment a user should send the amount
that will increase the balance to 100, settle on the blockchain, then send the
remaining amount.  In this scenario, Alice would send Charlie 5 units, settle
using a blockchain reverting their trustline balances to 0, then send the 
remaning 5 units for a remaining trustline balance of { Bob: -60, Charlie -5 }.

The blockchain your users will be using to settle their credit balances is
called FakeChain.  We have implemented FakeChain for you and outline how you can
interact with FakeChain through the endpoints below.

Implement a program that exposes an interactive command line interface for each
user participating in this network. Once a user starts the interactive prompt 
they should be able to send money to other users using a trustline, view their 
trustline and FakeChain balances, and settle their trustlines over FakeChain.

If you have any questions, please contact us in a group text and we'll get back 
to you as soon as possible.

Austin: ```(574) 849 - 9823```

Dino: ```(336) 391 - 1192```

### Constraints

- **Users in the network must be able to communicate from different computers**
- Each user keeps track of their own trustline balances
- Users must be able to handle multiple peers; there should be only one trustline per peer
- Before closing a session, a user must settle all outstanding debts on FakeChain 
  (do not worry about outstanding credits)
- We are looking for a decentralized network, trustlines should not be tracked or stored 
  remotely (i.e. on a server)
- One should be able to add new users into the network at will

Do not worry about writing tests, we are more interested in seeing your 
architecture decisions.  Also, do not worry about handling malicious nodes, you
can assume that all of the participants in the network will behave honestly.

### FakeChain REST API

##### Endpoint: 
This is the endpoint you will use to query FakeChain.
```http://ec2-34-222-59-29.us-west-2.compute.amazonaws.com:5000/```
The various queries and submissions you can make to FakeChain are outlined
below.

##### 

```/add_user```

*URL params*

**candidate**: access key given at top of prompt

**public_key**: node name

**amount**: initial starting funds

**private_key**: secret key to submit payments to FakeChain

**peering_info**: custom JSON object of your design containing information 
used to connect to other users in the network

```/delete_all_users```

*URL params:*

**candidate**: access key given at top of prompt

```/get_users```

*URL params*

**candidate**: access key given at top of prompt

```/pay_user```

*URL params*

**candidate**: access key given at top of prompt

**sender**: public_key of sending node

**receiver**: public_key of receiving node

**private_key**: private_key of sender to authorize payment

**amount**: amount to send

### Example Terminal Output

#### Alice

```
$ ./start-user ... # ( init options )
User Alice created and registered on FakeChain!
> open_trustline # ( Bob peering options )
Trustline with Bob started!
> open_trustline # ( Bill peering options )
Trustline with Bill rejected.
> pay Bob 10
Sent
> balance 
Bob: -10
Total: -10
> exit
Settling all trustlines.
Goodbye.
```

#### Bob

```sh
$ ./start-user ... # ( init options )
User Bob created and registered on FakeChain!
Alice wants to start a trustline, accept?
[Y/n]
Trustline with Alice opened!
> balance
Alice: 0
Total: 0
You were paid 10!
> balance
Alice: 10
Total: 10
> exit
Settling all trustlines.
Goodbye.
```

#### Bill

```sh
$ ./start-user ... # ( init options )
User Bill created and registered on FakeChain!
Alice wants to start a trustline, accept?
[Y/n]
Trustline with Alice denied!
> exit
No trustlines to settle.
Goodbye.
```
