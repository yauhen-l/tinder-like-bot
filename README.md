# tinder-like-bot

Installation
```
> go get github.com/yauhen-l/tinder-like-bot
> go install github.com/yauhen-l/tinder-like-bot
```

Usage example:
```
> $GOPATH/bin/tinder-like-bot -fb-token %facebook_access_token%
```
The easiest way to get %facebook_access_token% is to use link below, log in and then pick the auth token out of the URL you are redirected to:
https://www.facebook.com/dialog/oauth?client_id=464891386855067&redirect_uri=https://www.facebook.com/connect/login_success.html&scope=basic_info,email,public_profile,user_about_me,user_activities,user_birthday,user_education_history,user_friends,user_interests,user_likes,user_location,user_photos,user_relationship_details&response_type=token

Optional params and it's default values:
```
  -yes-limit 30          // script will stop when desired limit of likes reached
  -filter filter.json    // filter to like or pass
  -dry-run               // do not "like" or "pass" anyone - just log result
```

filter.json contents description:
```
{
    "ExcludeName": ["sam", "bob", "Anna", "Olya"],   // 'pass' if any matched case-insensitive
    "Schools": ["oxford", "University"],             // 'like' if any matched case-insensitive
    "CommonInterests": ["dog", "skiing"]             // 'like' if and matched case-insensitive
}
```