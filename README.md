# Powxy &ndash; anti-scraper reverse proxy

**Work in progress.**

Powxy uses a SHA-256 proof-of-work challenge to protect upstream resources from
scraper bots.

## Credits

Inspired by [Anubis](https://github.com/TecharoHQ/anubis). But much simpler.

## Bugs

- If a user is attempting to submit a POST request but their powxy cookie is
  invalid, powxy would redirect them to a challenge, and their POST data will
  be lost.
- It does not work when duplex connections are needed, e.g. with Git's Smart
  HTTP protocol.
- We should probably use
  [net/http/httputil.NewSingleHostReverseProxy](https://pkg.go.dev/net/http/httputil#NewSingleHostReverseProxy)
  instead of hand-wiring a half-broken proxy.
- We should probably add a JS/WASM version for browsers that support them.
- We should limit the request size.

## License

We are currently using the
[GNU Affero General Public License version 3](https://www.gnu.org/licenses/agpl-3.0.html).
