bearded-monitor
===============
Lean, fast, flexible monitoring

Usage
=====
```bash
$ ./bearded-monitor -p 50 -i 5s -s usr1 to_watch to_see
2015/03/27 16:08:48 Looking for commands containing: ["to_watch" "to_see"]
2015/03/27 16:08:48 Interval fixed to: 5s
2015/03/27 16:08:48 Looking for %CPU usage higher than 50%
2015/03/27 16:08:48 Kill signal to send: "user defined signal 1"

2015/03/27 16:08:49 Found 4 corresponding processes, with 0 > 50%.
...
```
