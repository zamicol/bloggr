![bloggr logo](bloggr_logo.png)


# bloggr
`bloggr` is a simple blog/article oriented Go application, designed to work with
`git`, that converts markdown files and associated images into static HTML
pages.`bloggr` works well with static site serving like Github.com. No
databases, no fancy HTTP servers.

`bloggr` runs a pipeline that converts an input directory with `.md` files into
finished `.html` files.  It also generates an article list that may be used
as the landing page for a blog.  The output from `bloggr` may be used as a stand
alone website or as apart of other generation pipelines.  Optionally, `bloggr`
also cryptographically hashes and signs files, which is ideal for lasting and
durable journalism and integrity if picked up by archival services..  (For
example, spez, [Reddit's CEO, edited people's
comments](https://www.theverge.com/2016/11/23/13739026/reddit-ceo-steve-huffman-edit-comments)
on the site.  Hashing and signing prevents such abuses.)

## Blog Files and Articles
 - An article may be a single file or contained in a subdirectories of the input
   directory. Subdirectories are helpful in organizing files, such as images,
   that may be associated with an article.
 - For each article, there is a `.md` file.  If a subdirectory does not contain
   a `.md` file, it is ignored.  If a subdirectory contains multiple `.md`
   files, it is ignored.  


## Header Mode 
For a `.md` file, the first line which starts with a markdown header, `#`
through `######`, is the title of the article and enters "header mode".  The
first empty, non-title line exits "header mode".

An input article must include a header line.  If a header line is not found, the
file is ignored by the article pipeline.

After the title, subheadings and metadata such as date, author, and tags may be
denoted in any `#` to `######` line.  The levels have no meaning (the levels
allow the selection of a desired HTML title size.), but we suggest using the
single `#` for title and `##` for subtitle.

## Footer Mode
The footer mode starts with a `---` line followed by a header line (`#`).

### Header/Footer Mode Labels
Inside of header mode, bloggr provides some pre-defined labels.  Authors may
define their own as desired. If a date or author isn't given, the current date
and default from config is used and field appearance is configurable in config.
 - `Date: ` 
 - `Author: `
 - `Tags: `

Multiple labels may appear on the same line. Date format may be YYYY/MM/DD,
YYYYMMDD, or a Unix timestamp. Tags may be multi-word and are delimited by a
comma.

```md
# Title
## Subtitle
### Date: 2025/1/1
#### Author: Zamicol
##### 2nd Subtitle
###### 3rd Subtitle
### Tags: Example Tag 1, Example Tag 2
```

## Example Blog
```md
# Gravity Wells are Sombreros
### Date: 20180921 Author: Zamicol
## Tags: Gravity, Special Relativity, Time Dilation

You've probably seen an illustration of a gravity well, something like this:

Or maybe you've seen one with Earth like this image:

[continued]

---
#### This article was original published on another platform.
```

## Processing Pipeline
### Directory Structure
`bloggr` on run looks for `config.json` in the current working directory and
processes the `blog` directory in accordance to the config.  Inside, the input
directory `0_markdown` is a staging area for articles before bloggr runs. The
input directory may contain stand alone `.md` files, representing stand alone
articles, or a sub-directory where each directory represents a single input
article. Other than the specified hierarchy, bloggr assumes a flat directory
structure and does not search for articles recursively. If a subdirectory does
not contain a `.md` file, or contains multiple `.md` files, it is ignored.

### HTML File Articles
Newly generated `index.html` files are created under a "article" directories,
`0_html` inside of a directory named using the article title. The
directory is named in title snake case with the prepended date. For example,
`0_html/20250114_Introduction_to_Coze/index.html` is generated for the
article `IntroductionToCoze.md`.

All other article files linked to from the article, such as images, are also
copied into an individual directory.

### Processed Files
After processing, source markdown files and related article files are moved to
`0_processed` as a directory,`0_processed/[article title]/index.md`, using title
snake case. For example, the title `#Introduction_to_Coze` results in a file
`0_processed/20250114_Introduction_to_Coze/index.md`.

### Example Directory Structure
When processing the blog directory, if `bloggr` files do not exist
they are created, such as `article_header.html`

#### Before Processing

Directories are denoted by square brackets `[]`.  
```
[blog]
	config.json
	random_file (ignored)
	incomplete_blog.md (ignored)
	coze_key.json
	[0_markdown]
		[Gravity Wells are Sombreros] (No date given and incorrect input format)
			GravityWells.md (file name/title mismatch)
			image1.png
			image2.png
		[Why_Information_Theory_is_Awesome_draft] (No date given and title/directory name mismatch)
			informationTheory.md (file name/title mismatch)
			image1.png
			image.png
		IntroductionToCoze.md (Title mismatch)
```

##### After Processing
```
[blog]
  config.json
  build.log
  random_file (ignored)
  incomplete_blog.md (ignored)
  article_header.html
  article_footer.html
  coze_key.json
  .gitignore
  [0_markdown]
    0_index.md    (Blog article list in markdown)
    [20250114_Introduction_to_Coze]
      index.md
    [20180921_Gravity_Wells_are_Sombreros]
      index.md
      image1.png
      image2.png
    [20250113_Why_Information_Theory_is_Awesome]
      index.md
      image1.png
      image2.png
  [0_html]
    index.html  (HTML blog main page, includes article list)
    [20250114_Introduction_to_Coze]
      .index.html
    [20250113_Why_Information_Theory_is_Awesome]
      index.html
      image1_SHA256_CU5nKqKeNXFPYddfIGoD3ZMFknpIzqlpZr7eXLdCNL8.png
      image2_SHA256_IO_az6B4WwW53t_CjIvQ7zM2Xk-hO_bfT0DbUz57LhA.png
    [20180921_Gravity Wells are Sombreros]
      index.html
      image1_SHA256_CU5nKqKeNXFPYddfIGoD3ZMFknpIzqlpZr7eXLdCNL8.png
      image2_SHA256_IO_az6B4WwW53t_CjIvQ7zM2Xk-hO_bfT0DbUz57LhA.png
```


On running bloggr, if the following files or directories don't exist, they are created.  

``` 
config.json
build.log
article_header.html
article_footer.html
coze_key.json
.gitignore
0_processed
0_processed/index.md
0_html
0_html/index.html
```

## Blog Main Page
`0_html/index.html` is the blog main page.  There is currently no pagination.

### Example Blog Main Page

```
Zamicol's Blog
My writings about tech, politics, information theory/cryptography, and life.

[2025/01/14] Introduction to Coze
[2025/01/13] Why Information Theory is Awesome (Tags: Cryptography)  
[2024/10/30] Gravity Wells are Sombreros (Tags: Gravity, Relativity, Time Dilation)
              

All rights reserved Zamicol
```

### Article List

`0_processed/index.md` is a list of articles generated by this
tool.  Article names are generated in so that they are easy to be sorted by latest
date.  Article names are generated in this order: 

The date, 
the title,
 if the following exist: 
 dash (-) followed by the subtitle
 Open parentheses:
  - author (optional)
  - tags   (optional)
 End parentheses



```
[2025/01/13] Why Information Theory is Awesome (Tags: Cryptography)  
```

Once `0_processed/index.md` is generated, it's html contents are inserted into
the `Article List` section of `0_html/index.html` 

So that html is manageable by hand, only the `Article List` section of
`0_html/index.html` is modified by bloggr. Articles are inserted between
the the `Article List` tags.

```html
<!-- Start Article List -->
<!-- End Article List -->
```

An example `index.html` file:

```
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
    <p><a href="[20250114_Introduction_to_Coze]/index.html">[2025/01/14] Introduction to Coze</a></p>
    <p><a href="[20250113_Why_Information_Theory_is_Awesome]/index.html">[2025/01/13] Why Information Theory is Awesome</a> (Author: Zamicol, Tags: Cryptography)</p>
    <p><a href="[20241030_Gravity_Wells_are_Sombreros]/index.html">[2024/10/30] Gravity Wells are Sombreros</a> (Author: Zamicol, Tags: Gravity, Relativity, Time Dilation)</p>
    <!-- End Article List -->
		<p> All rights reserved Zamicol</p>
</body>
</html>
```

## Article Header and Footer
An input `header.html` and `footer.html` is used to construct articles.  The
contents are inserted before and after the article content respectively.  

This may be used for header images, navigation, copyright section, and other
useful information that will be presented on each article.  

Example (default) Header:

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>{ARTICLE_TITLE}</title>
</head>
<body>
    <a href="/index.html">Back to blog home</a>
    <h1>{ARTICLE_TITLE}</h1>

```

Example Footer:

```html
    <p>By {ARTICLE_AUTHOR} on {ARTICLE_DATE}</p>
    <hr>
    <p>&copy; All rights reserved.</p>
</body>
</html>
```

## Reprocessing
To reprocess an article, delete the associated `0_html` subdirectory.

For example, delete `0_html/20250114_Introduction_to_Coze/index.html` run bloggr, which will catch `0_markdown/20250114_Introduction_to_Coze/index.md`

## History
A copy of all original markdown documents are kept in `0_markdown/[article_title/history/digest`

This allows editing and visible changes.

## Editing
Editing with history is supported. Simply edit the file. `0_markdown/20250114_Introduction_to_Coze/index.md` and rerun bloggr.  



### Content Index
For versioning and tracking articles, bloggr uses a content index.

`content_index.txt`

SHA256_CU5nKqKeNXFPYddfIGoD3ZMFknpIzqlpZr7eXLdCNL8 image1_SHA256_CU5nKqKeNXFPYddfIGoD3ZMFknpIzqlpZr7eXLdCNL8.png

## Use with External other Web servers or Git Hosting
You may use the output `0_html` in projects via a symlink. For example,
if your website is `Example.com`, you may symlink the `a` directory to the
website root, so articles appears as
`Example.com/20250114_Introduction_to_Coze`

You may also use git directly: https://gist.github.com/zamicol/f160d05dd22c9eb25031


## Cryptography 
### Hashing And Signing
Cryptographic hashing is used to integrity protect blogs, and critically,
independently of `git`. This makes it much more difficult for nefarious
organizations and nation state actors, like the CIA or Mossad, to change the
content of your blogs.  This is especially important when you have important
things, or sensitive things, to say. Integrity protection makes it much easier
for others to make a copy of your blog and insure that the contents have not
changed.  This also insures that archival services can be trusted with copies of
your articles into the far future.

This is especially important as civilization is using less print and more
digital content.  Digital content is easy for powerful nefarious forces to
manipulate in subtle ways, unlike books or newspaper articles that may be hidden
away on a shelf.

Signing is one further step to make sure that others cannot write content in
your name.  As long as your key remains private, it is impossible for other
masquerade their writing as yours.

Cryptographic features are optional, but are provided for high security
guarantees.

### Image Hashing
Before parsing the markdown into HTML, use SHA-256 to hash images.  After the
image name, an underscore, digest name, another underscore, and the URL safe, no
padding, base 64 (b64ut) value of the hash is inserted before the file
extension.  These values are append to or updated if already present.

For example

    image1.png

Changes to

    image1_SHA256_CU5nKqKeNXFPYddfIGoD3ZMFknpIzqlpZr7eXLdCNL8.png

In the article `.md` file, the names of the images are updated with this new
name.  

### Article Hashing
Once the image file names are replaced with names that include a digest, a
section denoting the digest for the article is appended or updated at the end of
the article. 

```md
<!-- Start Article Crypto -->
This article's markdown file is integrity protected:
  SHA256:ZMXz8zGrcC89-Tt2dAF_YUvQC0gmxN-nhIgK9MwtDrc
<!-- End Article Crypto -->
```

### Article Signing
Articles may also be signed using Coze.  

For example, the article with the digest `ZMXz8zGrcC89-Tt2dAF_YUvQC0gmxN-nhIgK9MwtDrc` may be signed like this:

```json
{
  "pay": {
    "a": "ZMXz8zGrcC89-Tt2dAF_YUvQC0gmxN-nhIgK9MwtDrc",
    "alg": "ES256",
    "iat": 1623132000,
    "tmb": "cLj8vsYtMBwYkzoFVZHBZo6SNL8wSdCIjCKAwXNuhOk",
    "typ": "bloggr/sign"
  },
  "sig": "lDhE9Mm5WY2yjO8rkcL_AF3vE2sPrNKQ_ZIt88bYtOJsFPY0GeGEr8-kwFagko7w4jTZqe5-6zvuXYUsrRfqlQ"
}
```

`.gitignore` is populated with `coze_key.json` to prevent pushing to remote git repositories.  

It is then used to update the Article Crypto section.

```md
<!-- Start Article Crypto -->
This article's markdown file and images are integrity protected:
  SHA256:ZMXz8zGrcC89-Tt2dAF_YUvQC0gmxN-nhIgK9MwtDrc

This article's markdown file is signed by the author using Coze:
```
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
<!-- End Article Crypto -->
```

The key `coze_key.json` is used to sign.  If a key is not provided, a new one is generated.  






## Config Options
Config is expected to be located in the directory that the command is executed
as `config.json`.

```
Path           - Blog directory path.    Default: `.`                Example: /home/docs/blog
Input          - Input directory name.   Default `$Path/0_markdown` 
Output         - Output directory name   Default `$Path/0_html` 
Digest         - Algorithm for hashing.  Default `SHA-256`
MaxLogSize     - Max size for `build.log` until logs are rolled over.  Default: 10 MB

DateFormat     - Date Format. May also be YYYYMMDD or YYYY-MM-DD. Default YYYYMMDD 
DefaultAuthor  - Default Author.  If an author isn't provided. Default: Anonymous Example: Zamicol

MoveProcessed  - Boolean - Move input to processed directory on completion. Default: true
InsertDate     - Boolean - Insert current date into the article if one isn't given.  Default: true
InsertAuthor   - Boolean - Insert author into article if one isn't given.  Default: true

TitleDelimiter - Character used to delimit words in file names.  Default: `_`

History        - History for the article.  Default: True

ArticleListShowDate       - Boolean - Show the date on the main blog page under before the article title. Default: true
ArticleListShowTags       - Boolean - Show tags on the main blog page under the respective article title and author. Default: true
ArticleListShowAuthor     - Boolean - Show the author on the main blog page after the respective article title. Default: False
HashImages                - Boolean - Hash images and replace image names with a reference to the hash.   Default: true
HashArticle               - Boolean - Hash the contents of the `.md` article and append/update article digest section.  Default: true
SignArticle               - Boolean - Sign the article using Coze, generate key as file `coze_key.json` if it doesn't exist.  Default: true
DeleteAndRebuild          - Boolean - Delete all files/directories in `0_html` and rebuild all articles in `/0_markdown`.  Default: false
```

## Logging

Logging is provided to denote run settings and ignored files/directories (e.g.,
"Skipping incomplete_blog.md: no title found").  At the time of run, logs are
printed to the command line and appended to `build.log`. `build.log` is kept to
10 MB.  


## Article Styles
The following can be managed with CSS styles. 

Here's some things that the authors have done in the default style.css
- Images are centered.
- There is an option for single images per line, and multiple images per line.
- On the main blog page, metadata like the author and tags is in smaller text.


## FAQ
### Duplicate Titles?
On run, it's assumed that a duplicate is just the reprocessing of a file.  The
existing directory is deleted and is replaced with the new article.

### Why hash files independent of `git`?
Much of `git` is still using SHA1, many `git` environments are not designed for
integrity protection.  Also, this allows copying and distribution independently
of `git`.

### Can I not date my articles?
NO. And shame on you! You should always date online writings.
