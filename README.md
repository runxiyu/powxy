# Powxy &ndash; anti-scraper reverse proxy

Powxy uses a SHA-256 proof-of-work challenge to protect upstream resources from
scraper bots.

## Demo

A demo instance is available at [forge.runxiyu.org](https://forge.runxiyu.org/).
I intend to remove this in the near future and set up an actual demo instance
on somewhere sensible; forge.runxiyu.org doesn't get much bot traffic so it's
a bit unnecessary.

## Bugs

- **The cryptography hasn't been audited.** I wrote this after a tiring school
  week and this definitely needs some review.
- We should allow Git clients and RSS readers.
- If a user is attempting to submit a POST request but their powxy cookie is
  invalid, powxy would redirect them to a challenge, and their POST data will
  be lost.
- It does not work when duplex connections are needed, e.g. with Git's Smart
  HTTP protocol.
- Currently we round times to the nearest week for persistence, but this could
  cause issues if a user completes the challenge at the end of a rounded week.

## Installation

You need a working Go installation. Then either

```
go install go.lindenii.runxiyu.org/powxy@latest
```

or

```
git clone ssh://forge.lindenii.runxiyu.org/powxy/:/repos/powxy/
cd powxy
go install
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
  -upstream string
    	destination url base to proxy to (default "http://127.0.0.1:8080")
```

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

## Credits

Inspired by [Anubis](https://github.com/TecharoHQ/anubis). But much simpler.

## License

We are currently using the
[GNU Affero General Public License version 3](https://www.gnu.org/licenses/agpl-3.0.html).
