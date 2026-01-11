# Future Features

## UI/UX Improvements

- [x] Make original bot locations marked with star instead of bigger circle
- [x] Add full current attempt reset button
- [x] Add new solution button for mobile
- [x] Stop scrolling and pinch zooming on all devices
- [x] Slower solution replay
- [x] Button to replay a solution (triangle/play button)
- [ ] Update how to play screen to focus on controls for mobile or desktop as appropriate
- [x] Close the drawer when playing back a solution

## Bug Fixes

- [x] Investigate rendering issues for iOS devices
- [x] Don't allow multi player to join solo room
- [x] Stop replaying solution when jumping to another solution
- [x] Solutions shouldn't persist across games (investigate)
- [x] Solo mode should count game correct if any current solution is correct
- [x] Revisit robot touch target size
- [x] On late player waiting room, doesn't show current player, and game board doesn't show current player. Seems like it's not connected to the server.
- [x] If host leaves, all other players get pushed into waiting room.
- [x] Player can see the game and click I'm finished even though they're not playing.

## New Functionality

- [x] Add client joining between games functionality
- [x] Only the room creator can start the game
- [ ] Sort bot solutions to minimize different bot sequences (partial order sorting, constraints)
- [ ] Replay of solution should highlight next bot to move when switching bots, with slight extra pause at that point
