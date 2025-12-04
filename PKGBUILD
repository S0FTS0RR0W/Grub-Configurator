# Maintainer: Your Name <youremail@domain.com>

pkgname=grub-configurator
pkgver=0.1.0
pkgrel=1
pkgdesc="A Fyne-based GUI for configuring GRUB."
arch=('x86_64')
url=""
license=('GPL3') # Please update with your actual license

# Runtime dependencies
depends=('grub' 'gtk3' 'webkit2gtk' 'lsb-release')

# Build-time dependencies
makedepends=('go')

# The source is the current directory
source=('.')
sha256sums=('SKIP')

build() {
	# The module path seems to be 'grub-configurator' from your main.go
	# The project root is copied by makepkg into $srcdir
	cd "$srcdir"
	go build -o "$pkgname" .
}

package() {
	install -Dm755 "$srcdir/$pkgname" "$pkgdir/usr/bin/$pkgname"
}