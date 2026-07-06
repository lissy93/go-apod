

<h1 align="center">Go-APOD</h1>

<p align="center">
  <i>A CORS-enabled, no-auth wrapper to NASA's Astronomy Picture of the Day </i><br>
  <b>Public API: <a href="https://apod.as93.net/">apod.as93.net</a></b><br><br>
  <img width="100" src="https://raw.githubusercontent.com/Lissy93/go-apod/master/static/assets/pwa/apple-touch-icon.png" />
</p>


<details>
<summary><b>Contents</b></summary>

- [API Usage](#api-usage)
  - [`/apod`](#apod)
  - [`/image`](#image)
- [Deployment](#deployment)
  - [Heroku](#heroku)
  - [Docker](#docker)
  - [Executable](#from-executable)
  - [From Source](#from-source)
- [Development](#building-locally)
  - [Project Commands](#commands)
  - [Configuration Options](#environmental-variables)
- [Frontend App](#app)
- [Contributing](#contributing)
- [License](#license)

</details>

---

## API Usage


### `/apod`

> Returns full JSON info about todays picture

**Example**

```
GET https://apod.as93.net/apod
```

**Response**

```json
{
  "date": "2022-06-20",
  "explanation": "There, just right of center, what is that? The surface of Mars keeps revealing new surprises with the recent discovery of finger-like rock spires. The small nearly-vertical rock outcrops were imaged last month by the robotic Curiosity rover on Mars. Although similar in size and shape to small snakes, the leading explanation for their origin is as conglomerations of small minerals left by water flowing through rock crevices. After these relatively dense minerals filled the crevices, they were left behind when the surrounding rock eroded away.  Famous rock outcrops on Earth with a similar origin are called hoodoos. NASA's Curiosity Rover continues to search for new signs of ancient water in Gale Crater on Mars, while also providing a geologic background important for future human exploration.   Explore Your Universe: Random APOD Generator",
  "hdurl": "https://apod.nasa.gov/apod/image/2206/MarsFingers_Curiosity_1338.jpg",
  "media_type": "image",
  "service_version": "v1",
  "title": "Rock Fingers on Mars",
  "url": "https://apod.nasa.gov/apod/image/2206/MarsFingers_Curiosity_960.jpg"
}
```

---

### `/image`

> Returns todays image

**Example**

```html
<img
  src="https://apod.as93.net/image"
  alt="Astronomy Picture of the Day"
  width="350"
/>
```

**Response**

<img src="https://apod.as93.net/image" alt="Astronomy Picture of the Day" width="350" />

---

## Deployment

> _Go-APOD can be self-hosted, either with Docker, via the 1-click Vercel or Heroku deployment, or by running the executable directly._<br>
> A NASA API Key is required, which you can sign up for at [api.nasa.gov](https://api.nasa.gov/).

### Vercel

[![Deploy with Vercel](https://vercel.com/button)](https://vercel.com/new/clone?repository-url=https%3A%2F%2Fgithub.com%2FLissy93%2Fgo-apod&env=NASA_API_KEY&envDescription=Your%20NASA%20API%20key.%20It's%20free%2C%20get%20it%20at%20https%3A%2F%2Fapi.nasa.gov&envLink=https%3A%2F%2Fapi.nasa.gov&project-name=apod&repository-name=go-apod&demo-title=Go-APOD&demo-description=A%20demo%20is%20published%20to%20apod.as93.net&demo-url=https%3A%2F%2Fapod.as93.net%2F&demo-image=https%3A%2F%2Fraw.githubusercontent.com%2FLissy93%2Fgo-apod%2Fmaster%2Fstatic%2Fassets%2Fpwa%2Fapple-touch-icon.png)

### Heroku

[![Deploy to Heroku](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy?template=https://github.com/Lissy93/go-apod)

### Docker
A multi-arch container is available on DockerHub, under [`lissy93/apod`](https://hub.docker.com/r/lissy93/apod), or GHCR  as [`ghcr.io/lissy93/go-apod`](https://github.com/Lissy93/go-apod/pkgs/container/go-apod).<br> Or, use this [`docker-compose.yml`](https://github.com/Lissy93/go-apod/blob/master/docker-compose.yml) template, and just populate with your API key and run `docker compose up`.

```bash
docker run -p 8080:8080 -e NASA_API_KEY='XXX' -d lissy93/apod
```

### From Executable

Each release has pre-compiled binaries attached for Windows, Mac and Linux, which can be run directly.
From the [Releases Page](https://github.com/Lissy93/go-apod/releases), download and extract the version for your system, then execute it with: `NASA_API_KEY='XXX' ./go-apod`

### From Source

See the [Building Locally](#building-locally) section below

---


## Building Locally

> If you haven't already done so, you'll need to [install Go Lang](https://go.dev/doc/install).<br>
> Then clone the repo `git clone https://github.com/Lissy93/go-apod.git && cd go-apod`


### Commands
- Run Directly > `go run .`
- Compile App > `go build`
- Run Tests > `go test`

### Environmental Variables

- `NASA_API_KEY` (Required) - Your API Key, you can sign up for one at [api.nasa.gov](https://api.nasa.gov/)
- `PORT` (Optional) - The port to start the web server on, defaults to `8080`
- `CORS_ALLOWED_ORIGINS` (Optional) - Comma-separated list of origins which can use the API, defaults to `*` / all
- `NASA_BASE_URL` (Optional) - The base URL for upstream GET requests, defaults to NASA's APOD API
- `CACHE_TTL` (Optional) - How long to cache NASA's response for, defaults to `15m`

---

## App

> The service also includes an optional simple web app, which can be used to show todays image and associated information from the API.

<p align="center">
  <a href="https://apod.as93.net">
  <img src="https://i.ibb.co/rvCfrbn/go-apod-screenshot.png" width="600" />
  </a>
</p>

---

## Contributing

Contributions of any kind are very welcome, and would be much appreciated :)
For Code of Conduct, see [Contributor Convent](https://www.contributor-covenant.org/version/2/1/code_of_conduct/).

To get started, fork the repo, make your changes, add, commit and push the code, then come back here to open a pull request. If you're new to GitHub or open source, [this guide](https://www.freecodecamp.org/news/how-to-make-your-first-pull-request-on-github-3#let-s-make-our-first-pull-request-) or the [git docs](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/creating-a-pull-request) may help you get started, but feel free to reach out if you need any support.

[![contributors](https://contrib.rocks/image?repo=lissy93/go-apod)](https://github.com/Lissy93/go-apod/graphs/contributors)

---

## License

> _**[Lissy93/Go-APOD](https://github.com/Lissy93/go-apod)** is licensed under [MIT](https://github.com/Lissy93/go-apod/blob/master/LICENSE) © [Alicia Sykes](https://aliciasykes.com) 2022._<br>
> <sup align="right">For information, see <a href="https://tldrlegal.com/license/mit-license">TLDR Legal > MIT</a></sup>

<details>
<summary>Expand License</summary>

```
The MIT License (MIT)
Copyright (c) Alicia Sykes <alicia@omg.com> 

Permission is hereby granted, free of charge, to any person obtaining a copy 
of this software and associated documentation files (the "Software"), to deal 
in the Software without restriction, including without limitation the rights 
to use, copy, modify, merge, publish, distribute, sub-license, and/or sell 
copies of the Software, and to permit persons to whom the Software is furnished 
to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included install 
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED,
INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANT ABILITY, FITNESS FOR A
PARTICULAR PURPOSE AND NON INFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
```

</details>

---

<!-- License + Copyright -->
<p  align="center">
  <i>© <a href="https://aliciasykes.com">Alicia Sykes</a> 2022 - Present</i><br>
  <i>Licensed under <a href="https://gist.github.com/Lissy93/143d2ee01ccc5c052a17">MIT</a></i><br>
  <a href="https://github.com/lissy93"><img src="https://cdn.as93.net/84m3gc?w=56" /></a><br>
  <sup>Thanks for visiting :)</sup>
</p>

<!-- Dinosaurs are Awesome -->
<!-- 
                        . - ~ ~ ~ - .
      ..     _      .-~               ~-.
     //|     \ `..~                      `.
    || |      }  }              /       \  \
(\   \\ \~^..'                 |         }  \
 \`.-~  o      /       }       |        /    \
 (__          |       /        |       /      `.
  `- - ~ ~ -._|      /_ - ~ ~ ^|      /- _      `.
              |     /          |     /     ~-.     ~- _
              |_____|          |_____|         ~ - . _ _~_-_
-->

