![bloggr logo](bloggr_logo.png)


# bloggr
`bloggr` is a simple blog/article oriented Go application, designed to work with
`git`, that converts markdown files and associated images into static HTML
pages.`bloggr` works well with static site serving like Github. No
databases, no fancy HTTP servers.

`bloggr` runs a pipeline that converts an input directory with `.md` files into
finished `.html` files.  It also generates an article list that may be used
as the landing page for a blog.  The output from `bloggr` may be used as a stand
alone website or as apart of other generation pipelines.  Optionally, `bloggr`
also cryptographically hashes and signs files, which is ideal for lasting and
durable journalism and integrity if picked up by archival services.  (For
example, spez, [Reddit's CEO, edited people's
comments](https://www.theverge.com/2016/11/23/13739026/reddit-ceo-steve-huffman-edit-comments)
on the site.  Hashing and signing prevents such abuses.)

## Directories, Blog Files, and Articles
 - An article may be a single file or contained in a subdirectory.
   Subdirectories are helpful in organizing files, such as images, that may be
   associated with an article.  Any file in an article subdirectory is also
   copied to the output, even if it is not referenced.
 - For each article, there is a `.md` file.

## Files and Directories
Defaults from the article path root. On run, if the following files or
directories don't exist, they are created.

 - `config.json`
 - `build.log`
 - `coz_key.json`
 - `.gitignore`
 - `markdown`
 - `markdown/index.md`
 - `public`
 - `public/article_header.html`
 - `public/article_footer.html`
 - `public/article`
 - `public/index.html`
 - `public/style.css`
 - `public/content/digests/[digest]`
 - `public/content/history.txt`
 - `public/content/index.txt`


### Directory Structure
**config.json** 
`bloggr` on run looks for <current working directory>/`config.json`  and
processes the `blog` directory in accordance to the config.  

**markdown**
Inside, the input directory `markdown` is a staging area for articles before
bloggr runs. The input directory may contain stand alone `.md` files,
representing stand alone articles, or a sub-directory where each directory
represents a single input article. Other than the specified hierarchy, bloggr
assumes a flat directory structure and does not search for articles recursively.
Articles are only searched for at depth 1. If a subdirectory does not contain a
`.md` file it is ignored, and if it contains multiple `.md` files it is skipped
and an error is logged.

**public**
New `index.html` files are created inside `public` in a sub-directory using the
article title, in title snake case with the prepended date. For example,
`public/article/20250114_Introduction_to_Coz/index.html` is generated for the
article `IntroductionToCoz.md`. All other article files, such as images, are
also copied into the subdirectory.

### Processing Pipeline

Key properties
- Processes articles in place.
- Directories are specified by `config.json` in the current working directory.

### Example Directory Structure
When processing the blog directory, if `bloggr` files do not exist
they are created, such as `article_header.html`

#### Before Processing

Directories are denoted by square brackets `[]`.  
```
[blog]
	config.json
	unrelated_file (ignored)
	coz_key.json
	[markdown]
		[Gravity Wells are Sombreros] (No date given and incorrect input format)
			GravityWells.md (file name/title mismatch)
			image1.png
			image2.png
		IntroductionToCoz.md (Title mismatch)
```

##### After Processing
```
[blog]
  config.json
  build.log
  unrelated_file  (ignored)
  coz_key.json
  .gitignore
  [markdown]
    index.md                    (Blog article list in markdown)
    [20250114_Introduction_to_Coz]
      index.md
    [20180921_Gravity_Wells_are_Sombreros]
      index.md
      image1_[content_hash].png
      image2_[content_hash].png
  [public]
    index.html                   (HTML blog main page, includes article list)
    article_header.html
    article_footer.html
    [article]
      [20250114_Introduction_to_Coz]
        index.html
      [20180921_Gravity_Wells_are_Sombreros]
        index.html
        image1_SHA-256:CU5nKqKeNXFPYddfIGoD3ZMFknpIzqlpZr7eXLdCNL8.png
        image2_SHA-256:IO_az6B4WwW53t_CjIvQ7zM2Xk-hO_bfT0DbUz57LhA.png
    [content]
      history.txt  // History of revisions
      index.txt    // Index of file names and digests.
      [digests]
        <digest>
```

## Metadata Mode - Header and Footer
There are two metadata modes: header mode and footer mode.
For a `.md` file, the first line which starts with a markdown header, `#`
through `######`, is the title of the article and enters "header mode".  

An input article must include a header line.  If a header line is not found, the
file is ignored and the article pipeline ignore the file/article directory.

After the title, subheadings and metadata such as date, author, and tags may be
denoted in any `#` to `######` line.  The levels have no meaning (the levels
allow the selection of a desired HTML title size.), but we suggest using the
single `#` for title and `##` for subtitle.

The footer mode starts with a `---` line followed by a header line.

The first empty, non-header line exits any metadata mode.

**Metadata Mode Labels**
Inside of metadata mode, bloggr provides some pre-defined labels.  Authors may
define their own as desired. If a date or author isn't given, the current date
and default from config is used and field appearance is configurable in config.
Author and tag are variadic and accept multiple values.

 - `Date: ` 
 - `Author: `
 - `Tag: `

Multiple labels may appear on the same line. Date format may be YYYY/MM/DD,
YYYYMMDD, or a Unix timestamp. Tags may be multi-word and are delimited by a
comma.

```md
# Title
## Subtitle
#### 2nd Subtitle
#### Date: 20250101, Author: Zamicol
#### Tag: Example Tag 1, Example Tag 2
```

## Example Blog
```md
# Gravity Wells are Sombreros
###### Date: 20180921, Author: Zamicol, Tag: Gravity, Special Relativity, Time Dilation
###### Original post published Friday, September 21, 2018.  Modified 20260925

You've certainly seen an illustration of a gravity well, something like this:

![gravity dip](gravity_dip.png)

[continued]
```

### Main Page and Article List
`public/index.html` is the blog main page and is derived from
`markdown/index.md`.  There is currently no pagination.

`markdown/index.md` is a list of articles generated by this tool.  Article
names are generated in so that they are easy to be sorted by latest date.
Article names are generated in this order: 

The date, 
the title,
 if the following exist: 
 dash (-) followed by the subtitle
 Open parentheses:
  - author (variadic, optional)
  - tag   (variadic, optional)
 End parentheses

Once `markdown/index.md` is generated, it's html contents are inserted into
the `Article List` section of `public/index.html` 

So that html is manageable by hand, only the `Article List` section of
`public/index.html` is modified by bloggr. Articles are inserted between
the the `Article List` tags.

```html
<!-- Start Article List -->
<!-- End Article List -->
```

An example `index.html` file:

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Zamicol's Blog</title>
</head>
<body>
    <h1>Zamicol's Blog</h1>
    <p>My writings about tech, politics, information theory/cryptography, and life.</p>
    <!-- Start Article List -->
    <p><a href="[20250114_Introduction_to_Coz]/index.html">[2026/09/21] Introduction to Coz</a></p>
    <p><a href="[20241030_Gravity_Wells_are_Sombreros]/index.html">[2024/10/30] Gravity Wells are Sombreros</a> (Tags: Gravity, Relativity, Time Dilation)</p>
    <!-- End Article List -->
		<p> All rights reserved Zamicol</p>
</body>
</html>
```

## Article Header and Footer
An input `header.html` and `footer.html` is used to construct articles.  The
contents are inserted before and after the article content respectively. This
may be used for header images, navigation, copyright section, and other useful
information that will be presented on each article.  

Example (default) Header:

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>{ARTICLE_TITLE}</title>
</head>
<body>
    <a href="{URI_BASE_PATH}/index.html">Back to home</a>
    <h1>{ARTICLE_TITLE}</h1>
    [article]
```


Example Footer:
````html
<p>By {ARTICLE_AUTHOR} on {ARTICLE_DATE}</p>
<p><a href="{URI_BASE_PATH}/content/digests/{ARTICLE_DIGEST}:{ARTICLE_DIGEST_ALG}">Original markdown</a></p>
<!-- Start Article Signature, template variable: {ARTICLE_COZ} -->
This article's markdown file and assets are integrity protected and signed by 
the author using Coz:
  ```json
{ARTICLE_COZ}
  ```

<h4>Article History</h4>
{ARTICLE_HISTORY}
<hr>
<p>&copy; All rights reserved.</p>
</body>
</html>
```

Which would populate as:
````html
<p>By Zamicol on 2026/09/21</p>
<p><a href="<URI_BASE_PATH>/content/digests/<digest>:SHA-256">Original markdown</a></p>
<!-- Start Article Signature, template variable: {ARTICLE_COZ} -->
This article's markdown file and assets are integrity protected and signed by 
the author using Coz:
  ```json
{
  "pay": {
    "article": "ZMXz8zGrcC89-Tt2dAF_YUvQC0gmxN-nhIgK9MwtDrc",
    "alg": "ES256",
    "iat": 1623132000,
    "tmb": "cLj8vsYtMBwYkzoFVZHBZo6SNL8wSdCIjCKAwXNuhOk",
    "typ": "bloggr/sign"
  },
  "sig": "lDhE9Mm5WY2yjO8rkcL_AF3vE2sPrNKQ_ZIt88bYtOJsFPY0GeGEr8-kwFagko7w4jTZqe5-6zvuXYUsrRfqlQ"
}
  ```

<h4>Article History</h4>
<ul>
  <li><a href="<URI_BASE_PATH>/content/history/1790000000_<digest>.html">Current 2026/09/21 (1790000000)</a></li>
  <li><a href="<URI_BASE_PATH>/content/history/1789900000_<digest>.html">2026/09/20 (1789900000)</a></li>
  <li><a href="<URI_BASE_PATH>/content/history/1789860000_<digest>.html">2026/09/19 (1789860000)</a></li>
</ul>
<hr>
<p>&copy; All rights reserved.</p>
</body>
</html>
````

## Editing and History
Bloggr supports editing with history. To edit, simply edit the source markdown
file. Bloggr uses file system time stamping (modification time) for efficient
processing. For example, edit `markdown/20250114_Introduction_to_Coz/index.md`
and rerun bloggr.

**History**
A copy of all original assets (documents and images) are kept in
`public/content/digests/[digest]`.  

`content/index.txt` stores original names.
```
CU5nKqKeNXFPYddfIGoD3ZMFknpIzqlpZr7eXLdCNL8:SHA-256 image1.png
rkkzQfAX69cERx-APub0cH1BmTX4mMqxFt704EgbL_o:SHA-256 gravity.md
```

History allows editing and visible changes while keeping the original. The
history index, `content/history.txt`:

```jsonl
{
  "path_input":"markdown/20250114_Introduction_to_Coz",        // Original markdown file name
  "path_output":"public/article/20250114_Introduction_to_Coz", // Output file name.

  "typ": "article",
  "id": "<initial digest>", // Article ID, based on initial digest
  "rev": "SHA-256:<digest>", // Current digest, derived from input md article.
  "prv": "SHA-256:<digest>", // Previous revision digest
  "now": 1790000000,
  "assets": {
    "coz_intermediary.png": "SHA-256:<digest>",
    "coz_is_small.png": "SHA-256:<digest>",
    "coz_logo.png": "SHA-256:<digest>",
  },

  // Metadata labels
  "title": "Introduction to Coz",
  "author": "Zamicol",
  "date": "20260921"    // Always in YYYYMMDD
}
```


**Content Digest Convention**
Normal digest names follow the [Coz naming
convention](https://github.com/Cyphrme/Coz#external-digest-serialization).
However, for the history directory, digests are store in reverse digest naming,
where the value comes first. This allows a shared name space for all digests.

CU5nKqKeNXFPYddfIGoD3ZMFknpIzqlpZr7eXLdCNL8:SHA-256

## Use with External other Web servers or Git Hosting
You may use the output `public` in projects via a symlink. For example, if your
website is `Example.com`, you may symlink the `a` directory to the website root,
so articles appears as `Example.com/20250114_Introduction_to_Coz`. You may also
use git directly via a "push-to-prod" git hook, for example:
https://gist.github.com/zamicol/f160d05dd22c9eb25031


## Cryptography 
### Hashing And Signing
Cryptographic features are optional. As civilization uses less print and more
digital content, hashing provides high security guarantees.  Independently of
`git`, cryptographic hashing is used to integrity protect articles, images, and
other assets. Digital content is easy for powerful nefarious forces to
manipulate in subtle ways, unlike books or newspaper articles that may be hidden
away on a shelf. Hashing protects your speech when you have important, or
sensitive, things to say, making it very difficult for nefarious organizations
and nation states.  Integrity protection helps backups and archives insure that
the contents have not changed, even into the far future.

Signing is one further step to make sure that others cannot write content in
your name.  As long as your key remains private, it is impossible for other
masquerade their writing as yours.

### Image Hashing
Before parsing the markdown into HTML, assets are hashed and "digest stamped".
After the base file name is an underscore, digest name, colon, and the b64ut
value of the hash is inserted before the file extension.  These values are
append to or updated if already present. In the article `.md` file, the names of
the images are updated with this new name.

For example, `image1.png` changes to

    image1_SHA-256:CU5nKqKeNXFPYddfIGoD3ZMFknpIzqlpZr7eXLdCNL8.png

### Article Hashing Signing
Once the image file names are replaced with names that include a digest, a
section denoting the digest for the article is appended or updated at the end of
the article. 

Articles, specifically the article digest, is signed using
[Coz](https://github.com/Cyphrme/Coz). For example, the article with the digest
`ZMXz8zGrcC89-Tt2dAF_YUvQC0gmxN-nhIgK9MwtDrc` are signed. The key `coz_key.json`
is used to sign.  If a key is not provided, a new one is generated.  
`.gitignore` is populated with `coz_key.json` to prevent pushing to remote git
repositories.  It is then used to update the Article Crypto section.


## Article Styles
The following can be managed with CSS styles. 

Here's some things that the authors have done in the default `style.css`
- Images are centered.
- There is an option for single images per line, and multiple images per line.
- On the main blog page, metadata like the author and tags is in smaller text.

## Logging
Logging is provided to denote run settings and ignored files/directories (e.g.,
"Skipping incomplete_blog.md: no title found").  At the time of run, logs are
printed to the command line and appended to `build.log`. `build.log` is kept to
10 MB.  

## Variables
List of variables available for the templates:

URI_BASE_PATH         - The URI base path for the `public` directory.  Example, `/articles`, which would point to `https://example.com/articles`
ARTICLE_TITLE         - The human readable title.  Example, `Introduction to Coz`
ARTICLE_AUTHOR        - Article author
ARTICLE_DATE          - The human date.  Example, `2026/09/21`
ARTICLE_DATE_YYYYMMDD - Always YYYYMMDD format date.
ARTICLE_DIGEST        - The markdown digest of this article. in form `<alg>:<digest>`
ARTICLE_DIGEST_VALUE  - The markdown digest of this article. in form `<digest>`
ARTICLE_DIGEST_ALG    - The markdown digest of this article. in form `<alg>`
ARTICLE_COZ           - The Coz signature HTML section for the article.
ARTICLE_HISTORY       - The simple HTML for the history section.



## Config Options
Config is expected to be located in the directory that the command is executed
as `config.json`.

```
URI_BASE_PATH  - The URI base path for the `public` directory.  Example: `/articles`
Path           - Blog directory path.    Default: `.`                Example: /home/docs/blog
Input          - Input directory name.   Default `$Path/markdown` 
Output         - Output directory name   Default `$Path/public` 
Digest         - Algorithm for hashing.  Default `SHA-256`
MaxLogSize     - Max size for `build.log` until logs are rolled over.  Default: 10 MB
DateFormat     - Date Format. May also be YYYYMMDD or YYYY-MM-DD. Default YYYYMMDD 
DefaultAuthor  - Default author if not provided. Default: Anonymous
TitleDelimiter - Character used to delimit words in file names.  Default: `_`

InsertDate     - Boolean - Insert current date into the article if one isn't given.  Default: true
InsertAuthor   - Boolean - Insert author into article if one isn't given.  Default: true
StyleCSS       - Boolean - Use `style.css`, creates if it doesn't exist.   Default: true
History        - Boolean - History for the article.  Default: true
HashImages     - Boolean - Hash images and replace image names with a reference to the hash.   Default: true
HashArticle    - Boolean - Hash the contents of the `.md` article and append/update article digest section.  Default: true
SignArticle    - Boolean - Sign the article using Coz, generate ES256 key as file `coz_key.json` if it doesn't exist.  Default: true

ArticleListShowDate     - Boolean - Show the date on the main blog page under before the article title. Default: true
ArticleListShowTags     - Boolean - Show tags on the main blog page under the respective article title and author. Default: true
ArticleListShowAuthor   - Boolean - Show the author on the main blog page after the respective article title. Default: False
DeleteAndRebuild        - Boolean - Delete all files/directories in `article` and rebuild all articles in `/markdown`.  Default: false
```


## FAQ
### Why hash files independent of `git`?
Much of `git` is still using SHA1, many `git` environments are not designed for
integrity protection.  Also, this allows copying and distribution independently
of `git`.  (We could use git object store directly, that is to interface with
git on a very low level, but we don't have time to do that now.)

### Can I not date my articles?
No. You should always date online writings. This drives me bonkers.

## See Also
See https://github.com/Cyphrme/Path for file path naming, specifically "file base".


## Appendix

### Sharding
In the future, if a directory or index becomes too large it is sharded based on
the b64 digest for the content index or date in the case of the history index.
The first b64 letter provides 64 shards (6 bits). However sharding is a pain for
static links, see the appendix.

### Watching for edits psuedo code
```psuedo code
if current.mtime != stored.mtime || current.size != stored.size {
    // treat as changed → re-hash / re-process
} else {
    // treat as unchanged → skip
}
```

### HTML History Links
On edit, the 

### Forced re-run
Although not suggested, reprocessing is also possible. Delete the associated
`public/article/[article]` subdirectory and re-run `bloggr`. For example, delete
`public/article20250114_Introduction_to_Coz/index.html` and run bloggr, which will
re-process `markdown/20250114_Introduction_to_Coz/index.md`. 



## See Also

https://github.com/Cyphrme/WatchMod
https://github.com/Cyphrme/Coz
https://github.com/Cyphrme/FileVer
https://github.com/Cyphrme/Path

### TODO
TODO think about a "purge" command.



