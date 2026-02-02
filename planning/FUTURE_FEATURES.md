# Future Features

## UI/UX Improvements

- [x] Make original bot locations marked with star instead of bigger circle
- [x] Add full current attempt reset button
- [x] Add new solution button for mobile
- [x] Stop scrolling and pinch zooming on all devices
- [x] Slower solution replay
- [x] Button to replay a solution (triangle/play button)
- [ ] Update how to play screen to focus on controls for mobile or desktop as appropriate
- [ ] When the solution drawer is closed, show brief indicators for each solution (e.g., length) rather than just the current one, with the active solution highlighted
- [x] Close the drawer when playing back a solution
- [x] Update location of roomid on mobile
- [x] Update solution spacing on mobile to match desktop
- [x] When a player is dropped from a game and they came back online, there should be some reasonable indication in their UI and a way to reconnect.
- [ ] A player should be able to leave and disconnect immediately
- [x] Solution bin should only solutions from the top three players plus bot

## Bug Fixes

- [ ] Solo mode disconnect timeout is too aggressive (30 seconds). Since no one is waiting, solo players should have a longer timeout (e.g., 30 minutes) to allow returning to a game after being distracted.
- [x] Investigate rendering issues for iOS devices
- [x] Don't allow multi player to join solo room
- [x] Stop replaying solution when jumping to another solution
- [x] Solutions shouldn't persist across games (investigate)
- [x] Solo mode should count game correct if any current solution is correct
- [x] Revisit robot touch target size
- [x] On late player waiting room, doesn't show current player, and game board doesn't show current player. Seems like it's not connected to the server.
- [x] If host leaves, all other players get pushed into waiting room.
- [x] Player can see the game and click I'm finished even though they're not playing.
- [x] Fix waiting room player colors

## New Functionality

- [x] Add client joining between games functionality
- [x] Only the room creator can start the game
- [x] Sort bot solutions to minimize different bot sequences (partial order sorting, constraints)
- [x] Auto-delete your worst solution
