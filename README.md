# Generate assets from templates

Picasso is a tool that can take a template and data, and generate images from them.
The templates that Picasso uses are written in HCL (HashiCorp Configuration Language) and define the input, output and design of the resulting image.

## Running the generator

To use the generator to generate an image from the template defined in `template.hcl` you can run the binary with the generate command and pass in the template using the `-t` or `--template` flag.

The resulting image will be written to the location specified using the `-o` or `--output` flag.

```shell
# picasso generate -t <template> -o <output>
picasso generate -t template.hcl -o output.png
```

An example of a generated image using a template from the [templates](https://github.com/eveld/picasso-templates) repository is this speaker card that was generated using the HashiTalks regional speaker card template.

![Example](https://github.com/eveld/picasso-templates/raw/master/hashitalks/regional/examples/speaker_1line.png "Speaker card example")

### Passing in variables

If you want to override variables from the command line, you can specify them using the `--var` arguments and supplying a key/value pair as `key=value` where key is the name of the variable and value is the value of the variable.

```shell
# picasso generate -t <template> -o <output> --var <key>=<value>
picasso generate -t template.hcl -o output.png --var title="Hello World!"
```

You can specify as many `--var` flags as you want.

### Reading variables from csv files

When you need to generate many different images from the same template, you can pass in variables through a .csv file.

Any variables that are passed in via the csv file will be replaced.

In the following file the `title` and `date` values are specified. The names of the columns in the csv file need to correspond with variables in the template.

```csv
title,date
Hello,01/01/2021
World,02/02/2021
```

And run the generate command with the `--csv` flag to pass in the path to the csv file.

```shell
# picasso generate -t <template> -o <output directory> --csv <csv file path>
picasso generate -t template.hcl -o "images/" --csv data.csv
```

In this case an image will be generated for each of the rows in the csv file, and the values of the variables `title` and `date` will be set to the corresponding values in the csv file for that row. The generated files will be placed in the specified output directory `images/` and named `output` followed by a random hash by default.

To override the title of the generated images from csv data, you can pass in the `--csv-var` flag to specify which field in the csv file should be used to name the file. The filename will still be followed by a random hash to prevent overwriting of duplicate images.

```shell
# picasso generate -t <template> -o <output directory> --csv <csv file path>
picasso generate -t template.hcl -o "images/" --csv data.csv --csv-var title
```

This will for example result in images being generated in the `images/` directory and named `Hello-asg443.png` and `World-kjgr33.png`.

## Templates

The main components of a template are `layers`. These layers can contain text, images or colors, and be resized and positioned anywhere in the resulting image.

## Text layer

A text layer represents a piece of text in the image.

```ruby
# Draw the text "Hello World!" at the coordinate 450,200 with Klavika Bold at 80pt size.
layer "text" "helloworld" {
  content = "Hello World!"
  x = 450
  y = 200
  size = 80
  font = "fonts/klavika/bold.ttf"
}
```

The text layer has the following configurable fields:

| Field | Type | Description |
| --- | --- | --- |
| content | string  | The content to draw |
| x | number | The x coordinate of the upper left corner in pixels |
| y | number | The y coordinate of the upper left corner in pixels |
| width | number | The max width of the layer in pixels |
| size | number | The font size in points |
| font | string | The path to the font |

## Image layer

An image layer represents an image inside of the resulting image.

```ruby
# Draw the dog image at the coordinate 160,350 and resize it to be 320 pixels wide.
layer "image" "dog" {
  content = "images/dog.jpg"
  x = 160
  y = 350
  width = 320
}
```

The image layer has the following configurable fields:

| Field | Type | Description |
| --- | --- | --- |
| content | string  | The contents of the image to draw |
| x | number | The x coordinate of the upper left corner in pixels |
| y | number | The y coordinate of the upper left corner in pixels |
| width | number | The max width of the layer in pixels |
| height | number | The max height of the layer in pixels |

## Output

To control the size of the resulting image, you can specify the output block.

```ruby
output "png" {
  width = 1600
  height = 900
}
```

Output has the following configurable fields:

| Field | Type | Description |
| --- | --- | --- |
| width | number | The width of the resulting image in pixels |
| height | number | The height of the resulting image in pixels |

## Variables

To pass variables to the template, that can be used to make the template more dynamic, you can use the variable blocks.

```ruby
variable "avatar" {
  type = "string"
  default = "images/avatar.jpg"
}
```

Variables can be of two types, currently `string` and `number` and have a default value that will be used when no value is passed in for that variable.

Variables have the following configurable fields:

| Field | Type | Description |
| --- | --- | --- |
| type | string | The type of the variable. Currently `string` or `number` |
| default | string/number | The default value of the variable |

To use a variable in another piece of the template, you can use interpolation syntax:

```ruby
variable "title" {
  type = "string"
  default = "Hello World!"
}

layer "text" "title" {
    content = "${title}"
}
```

This will replace the `${title}` with the variable value that is passed into the template, or `"Hello World!"` when no value is passed in.

