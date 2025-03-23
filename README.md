# Powxy &ndash; anti-scraper reverse proxy

**Work in progress.**

Powxy uses a SHA-256 proof-of-work challenge to protect upstream resources from
scraper bots.

## Demo

A demo instance is available at [forge.runxiyu.org](https://forge.runxiyu.org/).
I intend to remove this in the near future and set up an actual demo instance
on somewhere sensible; forge.runxiyu.org doesn't get much bot traffic so it's
a bit unnecessary.

## Credits

Inspired by [Anubis](https://github.com/TecharoHQ/anubis). But much simpler.

## Bugs

- The cryptography hasn't been audited. I wrote this after a tiring school week
  and this definitely needs some review.
- If a user is attempting to submit a POST request but their powxy cookie is
  invalid, powxy would redirect them to a challenge, and their POST data will
  be lost.
- It does not work when duplex connections are needed, e.g. with Git's Smart
  HTTP protocol.
- We should probably use
  [net/http/httputil.NewSingleHostReverseProxy](https://pkg.go.dev/net/http/httputil#NewSingleHostReverseProxy)
  instead of hand-wiring a half-broken proxy.

## Contributing

Create a branch that begins with `contrib/` and push via SSH.

```
git clone https://forge.lindenii.runxiyu.org/powxy/:/repos/powxy/
cd powxy
git checkout -b contrib/whatever
# edit and commit stuff
git push -u origin HEAD
```

## Support

[`#chat`](https://webirc.runxiyu.org/kiwiirc/#chat)
on
[irc.runxiyu.org](https://irc.runxiyu.org/).

## License

We are currently using the
[GNU Affero General Public License version 3](https://www.gnu.org/licenses/agpl-3.0.html).
