# Powxy &ndash; anti-scraper reverse proxy

Powxy is a reverse proxy that sits in front of your upstream service and
enforces a SHA-256 proof-of-work challenge. The goal is to discourage scraping,
as it becomes expensive for a client to perform en masse.

## Demo

A demo instance is available at [git.runxiyu.org](https://git.runxiyu.org/).
It may experience occasional outages.

## Bugs

- **The cryptography hasn't been audited.** I wrote this after a tiring school
  week and this definitely needs some review.
- The JavaScript challenge solver is very, very slow. The solver written in C
  as presented in the web interface is a few thousand times faster.
- We should allow Git clients and RSS readers.
- If a user is attempting to submit a POST request but their powxy cookie is
  invalid, powxy would redirect them to a challenge, and their POST data will
  be lost.
- It does not work when duplex connections are needed, e.g. with Git's Smart
  HTTP protocol.
- Currently we round times to the nearest week for persistence, but this could
  cause issues if a user completes the challenge at the end of a rounded week.
- SHA-256 is actually a pretty bad choice for proof-of-work. We should use
  something like bcrypt or scrypt or argon2.

## Build

You need a working Go installation.

```
git clone ssh://forge.lindenii.runxiyu.org/powxy/:/repos/powxy/
cd powxy
make
```

## Usage

```
Usage of ./powxy:
  -difficulty uint
    	leading zero bits required for the challenge (default 17)
  -listen string
    	address to listen on (default ":8081")
  -secondary
    	trust X-Forwarded-For headers
  -source string
    	url to the source code (default "https://https://forge.lindenii.runxiyu.org/powxy/:/repos/powxy/")
  -upstream string
    	destination url base to proxy to (default "http://127.0.0.1:8080")
```

## Mechanism

The client sends a request to the proxy. The proxy hashes information about the
client's IP address and user agent to generate an identifier. It checks whether
the client has a cookie containing an HMAC of the identifier. The connection is
forwarded to the upstream server if and only if the identifier matches and the
HMAC is valid.

Otherwise, the client is presented with a challenge, which asks them to find
a nonce that, when appended to the identifier, results in a SHA-256 hash that
begins with a certain number of zero bits. The client must solve the challenge
and submit it through an HTML form, which is then validated by the proxy. If
validation passes, the client is issued a cookie containing their identifier's
HMAC, and is redirected to request the page again, this time with the necessary
cookie to pass the validation.

JavaScript is provided to automatically solve the challenge without user
interaction. Clients that do not run JavaScript need to solve the challenge
externally, e.g. via the C program provided near the HTML form, and submit
their nonce manually.

## Contribute

Create a branch that begins with `contrib/` and push to
[the repo on our home forge](https://forge.lindenii.runxiyu.org/powxy/:/repos/powxy/)
via SSH.

```
git clone ssh://forge.lindenii.runxiyu.org/powxy/:/repos/powxy/
cd powxy
git checkout -b contrib/whatever
# edit and commit stuff
git push -u origin HEAD
```

## Support

[`#chat`](https://webirc.runxiyu.org/kiwiirc/#chat)
on
[irc.runxiyu.org](https://irc.runxiyu.org/).

## Similar software

* [Anubis](https://github.com/TecharoHQ/anubis)
  (the original inspiration that prompted me to write this)
* [PoW! Bot Deterrant](https://git.sequentialread.com/forest/pow-bot-deterrent)

## License

We use a FreeBSD-style license. See the `LICENSE` file for details.
