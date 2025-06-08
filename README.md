# GOTO - A simple SSH manager #

[![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](https://raw.githubusercontent.com/grafviktor/goto/master/LICENSE)
[![Codecov](https://codecov.io/gh/grafviktor/goto/branch/develop/graph/badge.svg?token=tTyTsuCvNb)](https://codecov.io/gh/grafviktor/goto)

This tiny app helps to maintain a list of SSH servers. Unlike PuTTY, it doesn't incorporate any connection logic, relying on the `ssh` utility which should be installed on your system.

Supported platforms: macOS, Linux, Windows.

## Goal ##

 기존의 프로그램은 방향키를 사용하여 SSH 서버를 선택한 후 엔터키를 입력하여 접속하는 방식만 지원했습니다. 본 프로젝트에서는 마우스 클릭을 통해 서버를 선택하고 접속할 수 있는 기능을 추가하여, 사용자의 편의성을 더욱 향상시켰습니다. 이로 인해 사용자들은 더 직관적이고 빠르게 SSH 서버에 접속할 수 있게 되었습니다. 하지만, 터미널 환경(PuTTY 등)에 따라 마우스 이벤트가 지원되지 않을 수 있습니다.


## Requirements ##

기술 스택 및 라이브러리
- Go 언어
- Bubble Tea: TUI 기반 UI 라이브러리
- os 패키지: ssh_config 파일 파싱
- Docker: 개발 환경 격리 및 배포

마우스 클릭 기능 세부 구현
- 'tea.MouseMsg`로 마우스 이벤트 처리
- 클릭 좌표 기반 항목 선택 (itemHeight, listOffset 사용)
- 클릭된 좌표를 화면의 항목 크기에 맞게 변환하여, 해당 항목을 선택하도록 구현
- 더블 클릭 감지: 두 번의 클릭 사이 시간 차이를 확인하여, 더블 클릭이 발생했을 때 해당 호스트에 대   한 SSH 연결을 시도
(더블클릭 기능 코드는 구현했으나 작동하지 않아 현재는 한 번 클릭으로 접속 가능한 상태)

### 실행 방법 ###

1. Repo 클론 및 디렉토리 이동
   git clone https://github.com/ts9744/2025-OSP.git
   cd 2025-OSP
   
2. Docker 이미지 빌드
   docker build \
  --build-arg VERSION=1.0.0 \
  --build-arg BRANCH=feature/Click_Function \
  -t final_2021040024:v1 \
  .

3. Docker 컨테이너 실행 및 진입
   docker run -it final_2021040024:v1 /bin/bash

4. 도구 버전 확인 (컨테이너 내부)
   git --version
   go version
   make --version
   gg -v
   
5. GO 수동 설치
   wget https://go.dev/dl/go1.22.3.linux-amd64.tar.gz
   
   A. 압축 해제 (Sudo Used)
   sudo tar -C /usr/local -xzf go1.22.3.linux-amd64.tar.gz

   B. 설치용폴더 생성, 압축해제 (Sudo Unused)
   mkdir -p $HOME/local
   tar -C $HOME/local -xzf go1.22.3.linux-amd64.tar.gz

   이후 환경 변수 설정
   echo 'export GOROOT=$HOME/local/go' >> ~/.bashrc
   echo 'export PATH=$GOROOT/bin:$PATH' >> ~/.bashrc
   source ~/.bashrc
   
   설치 확인
   go version

6. 프로그램 실행:
   go run main.go

7. 프로그램 종료 방법:
   프로그램 내에서 esc를 누른 뒤, y 입력 후 Enter
