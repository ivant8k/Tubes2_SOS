<!-- Back to Top Link-->
<a name="readme-top"></a>

<br />
<div align="center">
  <h1 align="center">Tugas Besar 2 IF2211 Strategi Algoritma</h1>

  <p align="center">
    <h3>Solver untuk Little Alchemy 2</h3>
    <h4>Menggunakan Kombinasi DFS dan BFS</h4>
    <h3><a href="https://github.com/ivant8k/Tubes2_SOS">Repositori</a></h3>
    <br/>
    <a href="https://github.com/ivant8k/Tubes2_SOS/issues">Report Bug</a>
    ·
    <a href="https://github.com/ivant8k/Tubes2_SOS/issues">Request Feature</a>
    <br>
    <br>

    [![MIT License][license-shield]][license-url]
  </p>
</div>

<!-- CONTRIBUTOR -->
<div align="center" id="contributor">
  <strong>
    <h3>Made By:</h3>
    <h3>Kelompok SOS</h3>
    <table align="center">
      <tr>
        <td>NIM</td>
        <td>Nama</td>
      </tr>
      <tr>
        <td>10123006</td>
        <td>Muhammad Naufal Rayhannida</td>
      </tr>
      <tr>
        <td>13523129</td>
        <td>Ivant Samuel Silaban </td>
      </tr>
            <tr>
        <td>13523164</td>
        <td>Muhammad Rizain Firdaus</td>
      </tr>
    </table>
  </strong>
  <br>
</div>




<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
    </li>
    <li>
      <a href="#getting-started-front-end">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#installation">Installation</a></li>
      </ul>
    </li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
  </ol>
</details>

## External Links

- [Spesifikasi](https://docs.google.com/document/d/1aQB5USxfUCBfHmYjKl2wV5WdMBzDEyojE5yxvBO3pvc/edit?usp=sharing)
- [QNA](https://docs.google.com/spreadsheets/d/1SVCNEBOYS0_eKShaHFIrx_5YVOg-V1uiBX-fAHpypxg/edit?usp=sharing)

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- ABOUT THE PROJECT -->
## About The Project
Proyek ini adalah solver untuk permainan *Little Alchemy 2*, yang bertujuan mencari resep elemen menggunakan algoritma pencarian. Kami mengimplementasikan:

- **DFS (Depth-First Search)**: Menelusuri satu cabang graf secara mendalam sebelum beralih ke cabang lain, efisien jika elemen ada di cabang awal.
- **BFS (Breadth-First Search)**: Menjelajahi simpul lapis demi lapis, menjamin jalur terpendek, efisien untuk graf dangkal.

Proyek ini juga mendukung pencarian banyak resep (*multi-recipe*) dengan pendekatan *multithreading* menggunakan Go, dengan analisis efisiensi untuk elemen seperti *Obsidian* dan *Beach*. Aplikasi ini memiliki frontend (Next.js) dan backend (Go), yang dapat dijalankan menggunakan Docker.


<p align="right">(<a href="#readme-top">back to top</a>)</p>


<!-- GETTING STARTED -->
## Getting Started

### Prerequisites

Bahasa Pemrograman: Go (versi 1.18 atau lebih baru).
- Sistem Operasi: Linux, Windows, atau macOS dengan Go terinstal.
- Instalasi Go:
    - Unduh dan instal dari https://golang.org/dl/.
    - Verifikasi dengan perintah: ``go version.``

<p align="right">(<a href="#readme-top">back to top</a>)</p>

### Installation

#### How to install and use this project (without docker)

1. Clone repository
   ```sh
    git clone https://github.com/ivant8k/Tubes2_SOS
    cd src
   ```
2. Untuk backend:
   ```sh
   cd backend
   go run server.go
   ```
3. Untuk frontend:
   ```sh
   cd frontend
   npm install
   npm run dev
   ``` 
<br>

<p align="right">(<a href="#readme-top">back to top</a>)</p>

#### How to install and use this project (with docker)

1. **Clone repositori**:
   ```sh
   git clone https://github.com/ivant8k/Tubes2_SOS
   cd Tubes2_SOS
   ```
2. **Jalankan dengan Docker Compose**:
   - Pastikan Anda berada di direktori root (yang berisi `compose.yml`).
   - Bangun dan jalankan container:
     ```sh
     docker compose up --build
     ```
   - Jika sudah pernah membangun container sebelumnya, cukup jalankan:
     ```sh
     docker compose up
     ```
3. **Akses aplikasi**:
   - Frontend tersedia di `http://localhost:3000`.
   - Backend berjalan pada port yang ditentukan di `compose.yml` (periksa file untuk detail).

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- FEATURES -->
## Features

### 1. Melakukan pencarian resep dengan Algoritma BFS.
### 2. Melakukan pencarian resep dengan Algoritma DFS.
### 3. Melakukan pencarian resep dengan Algoritma Bidirectional.
### 4. Melakukan pencarian multi resep untuk satu elemen.
### 5. Hasil pencarian resep dengan menggunakan graf.
### 6. Menggunakan website Fandom Little Alchemy 2 sebagai sumber data yang digunakan dalam scraping.
### 7. Pengguna dapat memasukkan input elemen, max recipes (untuk multirecipes), dan algoritma pencarian.
### 8. Docker supportneeded to reach the target

<p align="right">(<a href="#readme-top">back to top</a>)</p>


<!-- CONTRIBUTING -->
## Contributing

If you want to contribute or further develop the program, please fork this repository using the branch feature.  
Pull Request is **permited and warmly welcomed**

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- LICENSE -->
## License
Proyek ini dilisensikan di bawah MIT License.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<br>
<h3 align="center">THANK YOU!</h3>

<!-- MARKDOWN LINKS & IMAGES -->
[license-shield]: https://img.shields.io/badge/License-MIT-yellow
[license-url]: https://opensource.org/licenses/MIT
