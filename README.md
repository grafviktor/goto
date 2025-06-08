# GOTO - A simple SSH manager #

[![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](https://raw.githubusercontent.com/grafviktor/goto/master/LICENSE)
[![Codecov](https://codecov.io/gh/grafviktor/goto/branch/develop/graph/badge.svg?token=tTyTsuCvNb)](https://codecov.io/gh/grafviktor/goto)

This tiny app helps to maintain a list of SSH servers. Unlike PuTTY, it doesn't incorporate any connection logic, relying on the `ssh` utility which should be installed on your system.

Supported platforms: macOS, Linux, Windows.

## Goal ##

 기존의 프로그램은 방향키를 사용하여 SSH 서버를 선택한 후 엔터키를 입력하여 접속하는 방식만 지원했습니다. 하지만 이 프로그램에 마우스 클릭을 통해 서버를 선택하고 접속할 수 있는 기능을 추가하여, 사용자의 편의성을 더욱 향상시켰습니다. 이로 인해 사용자들은 더 직관적이고 빠르게 SSH 서버에 접속할 수 있게 되었습니다. 하지만 Putty와 같은 터미널에서 프로그램을 실행 시, 마우스 이벤트를 지원하지 않아 해당 기능 사용이 불가능합니다.


## Requirements ##

기술 스택 및 라이브러리
- Bubble Tea: CLI 애플리케이션을 위한 Go 라이브러리로, 사용자 인터페이스를 구성하는 데 사용
- OS 패키지: 로컬 파일 시스템에서 ssh_config 파일을 읽어오기 위한 패키지 사용

마우스 클릭 기능 구현
- tea.MouseMsg를 사용하여 마우스 클릭 이벤트를 처리
- tea.MouseLeft 타입의 클릭 이벤트를 감지하고, 사용자가 클릭한 좌표를 기반으로 해당 항목을 선택하도록 구현
클릭한 항목을 선택하기 위한 계산:
- 클릭된 좌표를 화면의 항목 크기에 맞게 변환하여, 해당 항목을 선택하도록 구현
- itemHeight와 listOffset을 사용하여 항목 크기와 리스트 오프셋을 고려해 클릭된 항목을 계산
- 더블 클릭 감지: 두 번의 클릭 사이 시간 차이를 확인하여, 더블 클릭이 발생했을 때 해당 호스트에 대한 SSH 연결을 시도
- time.Now()와 lastClickTime을 비교하여 더블 클릭 여부를 확인하고, 더블 클릭 시 SSH 연결을 시도하는 로직을 추가
- isDoubleClick 변수를 사용하여 두 번의 클릭이 500ms 이내로 발생했을 때만 더블 클릭으로 처리
(더블클릭 기능 코드는 구현했으나 작동하지 않아 현재는 한 번 클릭으로 접속 가능한 상태)

### 실행 방법 ###

1. git clone https://github.com/ts9744/2025-OSP.git
   cd 2025-OSP/cmd/goto
   
3. 아래 명령어를 실행해 Docker 이미지 빌드
   docker build \
  --build-arg VERSION=1.0.0 \
  --build-arg BRANCH=feature/Click_Function \
  -t final_2021040024:v1 \
  .

4. Docker 이미지가 생성되었다면, 다음 명령어를 통해 컨테이너 진입
   docker run -it final_2021040024:v1 /bin/bash
   컨테이너 안에서 다음 명령어들을 실행해 설치된 도구 버전 확인 가능
   git --version
   go version
   make --version
   gg -v
   
5. go 명령어 설치
   wget https://go.dev/dl/go1.22.3.linux-amd64.tar.gz
   
   a.  압축 해제, 관리자용
   sudo tar -C /usr/local -xzf go1.22.3.linux-amd64.tar.gz
   b. 일반 사용자일 경우 설치용폴더 생성, 압축해제
   mkdir -p $HOME/local
   tar -C $HOME/local -xzf go1.22.3.linux-amd64.tar.gz

   이후 환경 변수 설정
   echo 'export GOROOT=$HOME/local/go' >> ~/.bashrc
   echo 'export PATH=$GOROOT/bin:$PATH' >> ~/.bashrc
   source ~/.bashrc
   
   설치 확인
   go version

   실행
   go run main.go

6. 실행 종료 방법
   프로그램 내에서 esc를 누른 후 y와 enter를 입력하면 프로그램이 종료됩니다.

### Libraries and Tools ###

To run this project in Docker, you'll need the following **libraries and tools** installed in the container:

- **Docker** >= 20.10.0
- **Go** >= 1.22.x (for building the project)
- **Yaml** >= 2.2.1 (for configuration files)
- **Bubbletea** (for UI components) >= 0.19.0

**Docker Image Setup**: The Docker image includes all dependencies needed for this project to run smoothly. You can find the versions of libraries used in the Dockerfile.

For example, in the Docker container:
- Go: 1.22.x
- Yaml: 2.2.1
- Bubbletea: 0.19.0

### Docker Image Versions ###

This project uses **Docker** to isolate the application environment. The Dockerfile and container include:

- Go 1.22.x
- Yaml 2.2.1
- Bubbletea 0.19.0

**Note**: To ensure compatibility, these versions must be consistent between your local machine and the Docker container.

## How to Install & Run ##

### 1. **Docker Image Setup** ###

Assuming you have Docker installed, follow the steps below:

1. **Download the Docker image** for the project:
   ```bash
   docker pull ts9744/2025-osp
