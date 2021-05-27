# River SDK
This is the SDK which could be used by GUI/CLI clients to connect to River (https://river.im) server.
There is also an CLI for debugging and testing purpose in cmd/cli-river.

We have two versions of the SDK, one is the complete solution named 'prime' and the second one 'mini'
is useful when we don't need the full functionality.

# How TO Build for Mobiles
### Prepare ios project:
1. `git clone https://git.ronaksoft.com/river/ios.git`
2. `cd ./ios`
3. `git submodule update --init --remote`
4. `pod install`
5. `export RIVER_IOS_PATH=$(pwd)`   
6. build SDK for iOS:
   1. change directory sdk root folder
   2. run `RIVER_IOS_PATH=. sh build-ios-prime.sh`
   3. run `RIVER_IOS_PATH=. sh build-ios-mini.sh`

for Android:
run `RIVER_IOS_PATH=<root_folder_of_ios_project> sh build-android-prime.sh`



# Main Components
We have four important sub package, we call them controllers.
1. File Controller
2. Network Controller
3. Queue Controller
4. Sync Controller

