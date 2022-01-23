# River SDK
This is the SDK which could be used by GUI/CLI clients to connect to River (https://river.im) server.
There is also an CLI for debugging and testing purpose in cmd/cli-river.

We have two versions of the SDK, one is the complete solution named 'prime' and the second one 'mini'
is useful when we don't need the full functionality.

# How TO Build for Mobiles
### iOS
1. `git clone https://git.ronaksoft.com/river/ios.git`
2. `cd ./ios`
3. `git submodule update --init --remote`
4. `pod install`
5. `export RIVER_IOS_PATH=$(pwd)`   to make it permanent you can add it your rc file, or prepend the build command with correct env value.   
6. build SDK for iOS:
   1. change directory sdk root folder
   2. run `sh build-ios-prime.sh`
   3. run `sh build-ios-mini.sh`

### Android
1. `git clone https://github.com/ronaksoft/river-android-core`
2. `cd ./river-android-core`
3. `git submodule update --init --remote`
4. `export RIVER_ANDROID_PATH=$(pwd)` to make it permanent you can add it your rc file, or prepend the build command with correct env value.
5. build SDK for iOS:
   1. change directory sdk root folder
   2. run `sh build-android-prime.sh`




# Main Components
We have four important sub package, we call them controllers.
1. File Controller
2. Network Controller
3. Queue Controller
4. Sync Controller

