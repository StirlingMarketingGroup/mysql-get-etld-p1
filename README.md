# MySQL `get_etld_p1`

Essentially that stands for "Get Effective Top-Level Domain Plus 1", which means that this function returns what you'd expect to be the root domain from a URL. 

For example, this page I'm using right now to edit this README has the URL "https://github.com/StirlingMarketingGroup/mysql-get-etld-p1/edit/master/README.md", and if you looked at this, you'd want "github.com" and "https://dev.mysql.com/doc/refman/8.0/en/cast-functions.html" should return "mysql.com". I've written this extension so I can aggregate traffic sources without having issues like considering "www.google.com" and "google.com" as separate sources.

## Usage

> string  **get_etld_p1** ( string  `url` )

### Parameters
**url** - The URL string

### Return values
This function will return null if it failed to parse the string as a URL (errors might appear in the MySQL error log depending on the error), or a string if it was able to parse the URL correctly.

### Examples

These are some MySQL statements and their outputs.

```mysql
select`get_etld_p1`('http://a.very.complex-domain.co.uk:8080/foo/bar');-- 'complex-domain.co.uk'
select`get_etld_p1`('https://www.bbc.co.uk/');-- 'bbc.co.uk'
select`get_etld_p1`('https://github.com/StirlingMarketingGroup/');-- 'github.com'
select`get_etld_p1`('https://localhost:10000/index');-- 'localhost'
select`get_etld_p1`('android-app://com.google.android.gm');-- 'com.google.android.gm'
select`get_etld_p1`('example.test.domain.com');-- 'domain.com'
select`get_etld_p1`('postgres://user:pass@host.com:5432/path?k=v#f');-- 'host.com'
select`get_etld_p1`('exzvk.omsk.so-ups.ru');-- 'so-ups.ru'
select`get_etld_p1`('http://10.64.3.5/data_check/index.php?r=index/rawdatacheck');-- '10.64.3.5'
select`get_etld_p1`('not a domain');-- null
```

The parsing is based off of the Go package here https://godoc.org/golang.org/x/net/publicsuffix, which uses the public suffix list from https://publicsuffix.org/.

## Installation

This extension is written in Go instead of C or C++, so you'll to install Go if you haven't already https://github.com/golang/go/wiki/Ubuntu (make sure you grab the latest version instead of the one through `apt`).

Once Go is installed, the following should get you the rest of the way there.

1. make sure libmysqlclient-dev is installed
    `apt install libmysqlclient-dev`
2. Find your MySQL plugins dir, which is likely "/usr/lib/mysql/plugin" (or you can find it by running `select @@plugin_dir;` on your MySQL server)
3. Navigate to the folder where you cloned this repository to, and run the following (replacing the plugin path with your own)
    `go build -buildmode=c-shared -o get_etld_p1.so && cp get_etld_p1.so /usr/lib/mysql/plugin/get_etld_p1.so`
4. Then on your MySQL server, run this to declare the function in your current schema
    ``create function`get_etld_p1`returns string soname'get_etld_p1.so';``
