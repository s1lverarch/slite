# Maintainer: Silvercore <s1lv3rcore@proton.me>
pkgname=slite
pkgver=1.8.26
pkgrel=1
pkgdesc="Independent rootless container engine (proot-based) — no distrobox/podman required"
arch=('x86_64' 'aarch64')
url="https://github.com/s1lverarch/slite"
license=('GPL-3.0-only')
depends=('proot' 'curl' 'tar')
makedepends=('go' 'git')
source=("git+https://github.com/s1lverarch/slite.git#tag=v${pkgver}")
sha256sums=('SKIP')

build() {
    cd "$srcdir/slite"
    export CGO_ENABLED=0
    export GOFLAGS="-mod=mod"
    go build -ldflags "-s -w -X main.version=${pkgver}" -o slite ./cmd/slite
}

check() {
    cd "$srcdir/slite"
    go vet ./...
}

package() {
    cd "$srcdir/slite"
    install -Dm755 slite "$pkgdir/usr/bin/slite"
    install -Dm644 LICENSE "$pkgdir/usr/share/licenses/$pkgname/LICENSE"
    install -Dm644 README.md "$pkgdir/usr/share/doc/$pkgname/README.md"
}
