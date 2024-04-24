# Running the docker container 
```shell
docker run -d -p 12128:12128 --name excel2image guidoxie/excel2image:0.2
```

# Parameters
* format:
  Output file format (png jpeg jpg)
* quality:
  Output image quality (between 0 and 100) (default 94)
* width:
  Set screen width, note that this is used only as a guide line (default 1024)
* height:
  Set screen height (default is calculated from page content) (default 0)

# Example 
```shell
curl --location --request POST 'http://127.0.0.1:12128/api/upload?format=jpeg&width=1024&quality=100' --form 'file=@"test.xlsx"' --output test.jpeg
```