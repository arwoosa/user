# CHANGELOG 1.1.24
## Changes
- Statistics change key

# CHANGELOG 1.1.23
## Changes
- Statistics changes
- Show event scheduled count in GET auth

# CHANGELOG 1.1.22
## Changes
- Fix bug on incorect variable for searching non friends in recommended
- User friends rewilding activity status
- Update take me away status

# CHANGELOG 1.1.21
## Changes
- Fix bug on notification not sent on new user create

# CHANGELOG 1.1.20
## Changes
- Autoadd setting defaults to 0 when creating new account
- Pending status shown in recommended friends
- Filter recommended friends by name and username

# CHANGELOG 1.1.19
## Changes
- Update regex change password

# CHANGELOG 1.1.18
## Changes
- Username update regex, 60 days once allowed change
- Password check regex: POST register, PUT auth/change-password, POST forget-password
- User name check regex: POST register, PUT auth
- Recommended friends: other users who are not friends

# CHANGELOG 1.1.17
## Changes
- User friends notification (PENDING, ACCEPTED)

# CHANGELOG 1.1.16
## Changes
- User friends recommended (officially recommended friends, people who went together in an event)
- Auto accept friend request based on user setting

# CHANGELOG 1.1.15
## Changes
- GET users, GET users/statistics. Renamed 2 endpoints
- GET user/{userId}/friends. Retrieve other user friends

# CHANGELOG 1.1.14
## Changes
-  GET friends, POST friends, PUT friends/:userFriendId, DELETE friends/:userFriendId, /user/friends/recommended- Changes to friend

# CHANGELOG 1.1.13
## Changes
- world/ranking-feelings - default to EXPERIENCE_1, allow only EXPERIENCE_1 to EXPERIENCE_6

# CHANGELOG 1.1.12
## Changes
- GET user?name= Show 
- user-following routes to trigger count on delete, create and update

# CHANGELOG 1.1.11
## Changes
- GENERAL - user-following, auth routes remove trailing slashes 

# CHANGELOG 1.1.10
## Changes
- 4, 4.1, 4.3 - Removal of Auth Guards

# CHANGELOG 1.1.9
## Changes
- 4.3 - Top 5 Ranking (Feelings/Rewild)

# CHANGELOG 1.1.8
## Changes
-  4.0 - World Statistics

# CHANGELOG 1.1.7
## Changes
-  _O1, _O1a, _O1b, _O1c - Get user detail and badges
- 4.2e ~ 4.2h - Tab4 Friends

# CHANGELOG 1.1.6
## Changes
- Get OOSA users 3.3.2.1 ~ 3.3.2.1c, 3.1.1.4 ~ 3.1.1.4a

# CHANGELOG 1.1.5
## Changes
- User badges - Change structure

# CHANGELOG 1.1.4
## Changes
- Breathing points
- OOSA daily award breathing point

# CHANGELOG 1.1.3
## Changes
- Changes in validation errorlist
- Changes in unauthorized HTTP Code 401 Unauthorised
- Addition o N5 badges

# CHANGELOG 1.1.2
## Changes
- Bind to Facebook
- Statistics

# CHANGELOG 1.1.1
## Changes
- General changes to fit API Spec

# CHANGELOG 1.1.0
## Changes
User 
- Changes in settings key users_setting_is_visible_friends, users_setting_is_visible_statistics, users_setting_visibility_activity_summary, users_setting_friend_auto_add
- Update avatar 
- Retrieve badges earned by users

Notification
- Retrieve user notifications