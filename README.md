


![bloggr logo](bloggr_logo.png)




# bloggr
`bloggr` is a simple blog oriented Go application that converts markdown files and
associated images and parses it into an HTML pages.

`bloggr` is designed to work with git as a source management tool and with
static site serving like Github.com. 

`bloggr` runs in pipeline that converts an input directory with .md files into
finished ".html" files.  It also generates a blog article list that may be used
as the landing page for a blog.  

The output from `bloggr` may be used as a stand alone website or as apart of other
generation pipelines.  

## Blog Files and Articles
 - A single blog post is an article.
 - An article may be a single file or contained in a subdirectories of the input
   directory. Subdirectories are helpful in organizing files, such as images,
   that may be associated with an article.
 - For each article, there is a `.md` file.  If a subdirectory does not contain
   a `.md` file, it is ignored.  If a subdirectory contains multiple `.md`
   files, it is ignored.  


## Header Mode - Blog title, subtitles, and tags.  
For a `.md` file, the first line which starts with a single `#` is the title of
the blog and enters "header mode". A subtitle may follow a title starting with a
`##`. There may be other lines before the header.  The first non-empty,
non-title line exits "header mode".  

An input article must include a `#` title line.  If no `#` title line is found,
the file is ignored.

`###` to `######` may be used for tags or other subtitles.  

Metadata such as date, author, and tags, may be denoted in any `##` to `######`
line.  The levels have no meaning, but we suggest using `##` for subtitle.  The
levels just allow the user to choose their desired HTML title size.  

The labels are 
 - `Date: ` 
 - `Author: `
 - `Tags: `

Date may be in YYYY/MM/DD or YYYYMMDD format.  
Tags may be multi-word and are delimited by a comma.  

```md
# Title
## Subtitle
### Date: 2025/1/1
#### Author: Zamicol
##### 2nd Sub Title
###### 3rd Sub Title
### Tags: Example Tag 1, Example Tag 2
```

## Example Blog

```
This was originally published at blog.zamicol.com
# Gravity Wells are Sombreros
### Date: 20180921
#### Author: Zamicol
## Tags: Gravity, Relativity, Time Dilation


You've probably seen an illustration of a gravity well, something like this:
Or maybe you've seen one with Earth like this image:

[continued]
```


## Processing and Directory Structure
### Example Directory Structure
Directories are denoted by square brackets `[]`.  Other than the hierarchies
denoted, the tool assumes a flat directory structure for articles and does not
search for new articles recursively.  

In the "blog" directory, there is an input directory named `0_to_publish`.
`0_to_publish` is a staging area for before when the tool runs. The input
directory may contain stand alone `.md` files representing an article or
sub-directories where each directory represents a single input article.  If a
subdirectory does not contain a `.md` file, or contains multiple `.md` files, it
is ignored.  

After processing, directories and stand alone `.md` articles are moved to
`0_processed`.  The respective article file name or the subdirectory name in the
`0_processed` directory, `0_processed/[file_name].md` or `0_processed/[article
title]/index.md`, is renamed to the article title using title snake case. Title
is used regardless of the original filename.  The date is also prepended to the
title.  `Introduction_to_Coze.md` becomes `20250114_Introduction_to_Coze.md` so
the directory is easily sorted by date.  

For example, the title `# Why Information Theory is Awesome` results in a file
named `Why_Information_Theory_is_Awesome.md` if it is a stand alone file, or
`0_processed/Why_Information_Theory_is_Awesome/index.md` if it is a directory. 

Newly generated `index.html` files are created under `0_published/a` inside of a
directory named based on the article title.  For example,
`0_published/a/20250114_Introduction_to_Coze/index.html` is generated for 



#### Before Processing
```
[blog]
	config.json
	random_file (ignore)
	incomplete_blog.md (ignore)
	coze_key.json
	[0_to_publish]
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
	random_file (ignore)
	incomplete_blog.md (ignore)
	article_header.html
	article_footer.html
	coze_key.json
	.gitignore
	[0_to_publish]
	[0_published]
		[0_processed]
			index.md    (Blog main page in markdown)
			20250114_Introduction_to_Coze.md
			[20180921_Gravity_Wells_are_Sombreros]
				index.md
				image1.png
				image2.png
			[20250113_Why_Information_Theory_is_Awesome]
				index.md
				image1.png
				image.png
		index.html  (HTML blog main page)
		[a]
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




On running bloggr, if the following files or directories don't exist, create them.  

``` 
config.json
build.log
article_header.html
article_footer.html
coze_key.json
.gitignore
0_to_publish 
0_published
0_published/index.html
0_published/a
0_published/0_processed
0_published/0_processed/index.md
```



## Blog Main Page
`0_published/index.html` is the blog main page.  

### Example Blog Main Page

```
Zamicol's Blog
My writings about tech, politics, information theory/cryptography, and life.

[2025/01/14] Introduction to Coze
[2025/01/13] Why Information Theory is Awesome (Author: Zamicol, Tags: Cryptography)  
[2024/10/30] Gravity Wells are Sombreros (Author: Zamicol, Tags: Gravity, Relativity, Time Dilation)
              

All rights reserved Zamicol
```

### Article List

`0_published/0_processed/index.md` is a list of articles generated by this
tool.Article names are generated in so that they are easy to be sorted by latest
date.  Article names are generated in this order: The date, the title, the
author, and tags.  If a date isn't given, the current date is ued. If an author
isn't given, the default author from the config is used.  For example:

```
[2025/01/13] Why Information Theory is Awesome (Author: Zamicol, Tags: Cryptography)  
```

Once `0_published/0_processed/index.md` is generated, it is then used to update
`0_published/index.html`. So that static headers and footers are easier to
create once by hand, only the `Article List` section of `0_published/index.html` is
modified by bloggr.


Once the .md file is generated, it's html contents are inserted into the
`Article List` section of `0_published/index.html`

Blogs are inserted between the the `Article List` tags.

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



# Article Header and Footer
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


## Cryptographic Hashing And Signing
Cryptographic hashing is used to integrity protect blogs independently of `git`.
This makes it much more difficult for nefarious organizations, like the NSA,
CIA, or Mossad, to change the content of your blogs.  This is especially
important when you have important things, or sensitive things, to say. Integrity
protection makes it much easier for others to make a copy of your blog and
insure that the contents have not changed.  

This is especially important as civilization is using less print and more
digital content.  Digital content is easy for powerful nefarious forces to
manipulate in subtle ways, unlike books or newspaper articles that may be hidden
away on a shelf.

Signing is one further step to make sure that others cannot write content in
your name.  As long as your key remains private, it is impossible for other
masquerade their writing as yours.  

Cryptographic features are optional, but are provided for higher security
conscious deployments.  

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
This article's markdown file is integrity protected:
  SHA256:ZMXz8zGrcC89-Tt2dAF_YUvQC0gmxN-nhIgK9MwtDrc

	This article's markdown file is signed by the author using Coze:
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
<!-- End Article Crypto -->
```

The key `coze_key.json` is used to sign.  If a key is not provided, a new one is generated.  






## Config Options
Config is expected to be located in the directory that the command is executed
as `config.json`.

```
Path           - Blog directory path.    Default: `.`                Example: /home/docs/blog 
Input          - Input directory name.   Default `0_to_publish` 
Output         - Output directory name   Default `0_published` 
Digest         - Algorithm for hashing.  Default `SHA-256`
MaxLogSize     - Max size for `build.log` until logs are rolled over.  Default: 10 MB

DateFormat     - Date Format. May also be YYYYMMDD or YYYY-MM-DD. Default YYYY/MM/DD 
DefaultAuthor  - Default Author.  If an author isn't provided. Default: Anonymous Example: Zamicol 

MoveProcessed  - Boolean - Move input to processed directory on completion. Default: true 
InsertDate     - Boolean - Insert a date on an article if one isn't given.  Default: true 
InsertAuthor   - Boolean - Insert author into article if one isn't given.  Default: true

TitleDelimiter - Character used to delimit words in file names.  Default: `_`

ArticleListShowDate       - Boolean - Show the date on the main blog page under before the article title. Default: true
ArticleListShowTags       - Boolean - Show tags on the main blog page under the respective article title and author. Default: true
ArticleListShowAuthor     - Boolean - Show the author on the main blog page after the respective article title. Default: true
HashImages                - Boolean - Hash images and replace image names with a reference to the hash.   Default: true
HashArticle               - Boolean - Hash the contents of the `.md` article and append/update article digest section.  Default: true
SignArticle               - Boolean - Sign the article using Coze, generate key as file `coze_key.json` if it doesn't exist.  Default: true
DeleteAndRebuild          - Boolean - Delete all files/directories in `a` and rebuild all articles based on `0_published/0_processed`.  Default: false
```

## Logging

Logging is provided to denote run settings and ignored files/directories (e.g.,
"Skipping incomplete_blog.md: no title found").  At the time of run, logs are
printed to the command line and appended to `build.log`. `build.log` is kept to
10 MB.  


## Post Styles
The following can be managed with CSS styles.  

Here's some things that the authors have done in the default style.css
- Images are centered.  
- There is an option for single images per line, and multiple images per line. 
- On the main blog page, metadata like the author and tags is in smaller text.  


## FAQ
### Why move the input to published? 
So that the entire blog is easy add to git, while keeping the input potentially
private unless it's been published.  

I personally add only `0_published` to a publish github repo while keeping the larger `blog` private.  

### Duplicate Titles?
On run, it's assumed that a duplicate is just the reprocessing of a file.  The
existing directory is deleted and is replaced with the new article.  

### Why hash files independent of `git`?
Much of `git` is still using SHA1, many `git` environments are not designed for integrity protection.  Also, this allows copying and distribution independently of `git`. 

### Why are articles put into a `a` directory?
This makes it easy to regenerate from `0_processed`.  `a` is short for "articles".  
