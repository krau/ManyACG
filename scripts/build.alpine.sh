while read line; do
  pkg="${line%%-[0-9]*}"
  ver="${line#${pkg}-}"
  echo "${pkg}=${ver}"
done < alpinepkgs.v.txt | xargs apk add --no-cache

wget  https://github.com/openSUSE/libeconf/archive/refs/tags/v0.7.10.tar.gz 
meson setup build -Ddefault_library=static
cd build
meson compile
meson install
nano /usr/lib/pkgconfig/mount.pc
# 修改 Requires.private 将 libeconf 添加到后面
# 原: Requires.private: blkid
# 改: Requires.private: blkid libeconf

wget https://download.osgeo.org/libtiff/tiff-4.7.1rc1.tar.gz
tar xvf tiff-4.7.1rc1.tar.gz
cd tiff-4.7.1
./autogen.sh
./configure --enable-static --disable-shared --prefix=/usr/local
make -j
make install

wget https://gitlab.freedesktop.org/xorg/lib/libxau/-/archive/libXau-1.0.12/libxau-libXau-1.0.12.tar.gz
tar xvf libxau-libXau-1.0.12.tar.gz
cd libxau
./autogen.sh
./configure --enable-static --disable-shared --prefix=/usr/local
make -j
make install

wget https://github.com/webmproject/libwebp/archive/refs/tags/v1.6.0.tar.gz -O libwebp-1.6.0.tar.gz
tar xf libwebp-1.6.0.tar.gz
cd libwebp-1.6.0
./autogen.sh
./configure --enable-static --disable-shared --with-pic
make -j
make install

# TODO: libheif support , x265 needs old cmake.
# while libavif is better, libvips does not support it yet.
# wget https://github.com/strukturag/libheif/releases/download/v1.20.2/libheif-1.20.2.tar.gz
# tar xf libheif-1.20.2.tar.gz
# cd libheif-1.20.2
# mkdir build && cd build
# cmake --preset=release-noplugins .. -DBUILD_SHARED_LIBS=OFF -DCMAKE_INSTALL_PREFIX=/usr/local -DENABLE_PLUGIN_LOADING=OFF 
# make -j
# make install
# wget https://bitbucket.org/multicoreware/x265_git/downloads/x265_3.6.tar.gz

wget https://github.com/libvips/libvips/releases/download/v8.17.2/vips-8.17.2.tar.xz
tar xf vips-8.17.2.tar.xz
cd vips-8.17.2
# 关闭不需要的构建
# sed -i -E "s/^(subdir\('(tools|test|fuzz)'\))/# \1/" "$FILE"
# # subdir('tools')
# # subdir('test')
# # subdir('fuzz')
meson setup build --default-library=static --prefix=/usr/local -Dmodules=disabled -Dwebp=enabled -Dheif=disabled -Dprefer_static=true -Dexamples=false
cd build
meson compile
meson install

cd /ManyACG

builtAt="$(date +'%F %T %z')"
gitCommit=$(git log --pretty=format:"%h" -1)
version=$(git describe --abbrev=0 --tags)

versionFlags="-X 'github.com/krau/ManyACG/internal/common.BuildTime=$builtAt' \
-X 'github.com/krau/ManyACG/internal/common.Commit=$gitCommit' \
-X 'github.com/krau/ManyACG/internal/common.Version=$version'"

vipsFlags=$(pkg-config --static --libs vips)

# nodynamic tag is for https://github.com/gen2brain/avif
CGO_ENABLED=1 go build \
    -tags nodynamic,netgo \
    -ldflags "-s -w $versionFlags -linkmode external -extldflags \"-static $vipsFlags\"" \
    -o manyacg