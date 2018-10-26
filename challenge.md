# Strata Labs Technical Challenge

## Logistics
This challenge is intended to be done at home and should take 1-4 hours. You should complete it in whatever programming language you feel most comfortable in and you may use any resources you like.

We will schedule a review session afterwards where we will ask you to discuss your solution and explain your design choices.

Upon completion, please send a zip of all the files used in your solution to info@stratalabs.io.

_Please do not share this challenge or your solution to it._

## Challenge
A “trustline” is a way to keep track of debt between two parties. The trustline balance starts at 0 and both parties independently track their views of the balance. If Alice owes Bob $10, then Alice sees a balance of -10 and Bob sees a balance of 10. If Alice sends Bob 10 more, then her balance would be -20, and Bob’s balance would be 20.

Upon opening a trustline with another node, they should agree on a max balance the trustline should reach. Trustlines should never exceed this amount, any payment over this limit must be settled on FakeChain. For example, if the max amount in a trustline is 100 and Alice sends a payment of size 250, then 50 should remain on the trustline and the rest of the payment should be sent over FakeChain.

Write a program that implements and exposes an interactive command line interface for each user participating in this network. Once a user starts the interactive prompt, they should be able to send money to other users and view their trustline balances as well as their funds on FakeChain.

If you have any questions, feel free to contact us and we'll get back to you as soon as possible.
Austin: (574) 849 - 9823
Dino: (336) 391 - 1192

### Constraints

- Each user keeps track of their own trustline balances
- Users must be able to handle multiple peers; there should only be one trustline per peer
- Assume the users are on different computers
- Before closing a session, a user must settle all outstanding debts on FakeChain (do not worry about outstanding credits)
- Trustlines should not be tracked or stored remotely (i.e. on a server)
- One should be able to add new users into the network at will

Do not worry about writing tests, we are more interested in seeing your architecture decisions.

### FakeChain REST API

##### Endpoint: http://ec2-34-222-59-29.us-west-2.compute.amazonaws.com:5000/
&nbsp;
##### /add_user
*URL params*
**candidate**: access key given at top of prompt
**public_key**: node name
**amount**: initial starting funds
**private_key**: secret key to authorize payments

##### /delete_all_users
*URL params:*
**candidate**: access key given at top of prompt

##### /get_users
*URL params*
**candidate**: access key given at top of prompt

##### /pay_user
*URL params*
**candidate**: access key given at top of prompt
**sender**: public_key of sending node
**receiver**: public_key of receiving node
**private_key**: private_key of sender to authorize payment
**amount**: amount to send

### Example Terminal Output

#### Alice

```sh
$ ./start-user ... # ( init options )
User Alice created and registered on FakeChain!
> open_trustlines # ( peering options )
Trustline with Bob started!
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
Alice wants to start a trustline with max 100, accept?
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
Alice wants to start a trustline with max 100, accept?
[Y/n]
Trustline with Alice denied!
> exit
No trustlines to settle.
Goodbye.
```
