# github-console

## Install

```shell script
go get -v github.con/zhcppy/github-console
```

## Use

* Pressing the tag key will prompt or complete.
* Entering 'exit/ctrl-D' closes the console.
* Will create a .gh_history file in the $HOME directory to save used commands.
 
```
➜  github-console git:(master) github-console
Welcome to the Github console!
$🐌 Users.Get(golang)
01 - github.User:
{
  "avatar_url": "https://avatars3.githubusercontent.com/u/4314092?v=4",
  "blog": "https://golang.org",
  "created_at": "2013-05-01T18:00:52Z",
  "events_url": "https://api.github.com/users/golang/events{/privacy}",
  "followers": 0,
  "followers_url": "https://api.github.com/users/golang/followers",
  "following": 0,
  "following_url": "https://api.github.com/users/golang/following{/other_user}",
  "gists_url": "https://api.github.com/users/golang/gists{/gist_id}",
  "gravatar_id": "",
  "html_url": "https://github.com/golang",
  "id": 4314092,
  "login": "golang",
  "name": "Go",
  "node_id": "MDEyOk9yZ2FuaXphdGlvbjQzMTQwOTI=",
  "organizations_url": "https://api.github.com/users/golang/orgs",
  "public_gists": 0,
  "public_repos": 48,
  "received_events_url": "https://api.github.com/users/golang/received_events",
  "repos_url": "https://api.github.com/users/golang/repos",
  "site_admin": false,
  "starred_url": "https://api.github.com/users/golang/starred{/owner}{/repo}",
  "subscriptions_url": "https://api.github.com/users/golang/subscriptions",
  "type": "Organization",
  "updated_at": "2019-12-06T03:53:43Z",
  "url": "https://api.github.com/users/golang"
}
```