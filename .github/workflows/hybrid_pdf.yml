name: Hybrid Wiki PDF Export

on:
  workflow_dispatch: # 支持手动点击按钮触发

jobs:
  build:
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout Code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Setup Python
        uses: actions/setup-python@v5
        with:
          python-version: '3.11'

      - name: Install Chrome and Dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y google-chrome-stable
          pip install -r python_finisher/requirements.txt

      - name: Execute Go Crawler
        run: |
          cd go_crawler
          # 自动解决 go.mod 和 go.sum 依赖问题
          rm -f go.mod go.sum
          go mod init wiki-crawler
          go get github.com/chromedp/chromedp
          go get github.com/chromedp/cdproto/page
          go mod tidy
          go run main.go

      - name: Execute Python Finisher
        run: |
          # Go 运行后会在根目录产出 metadata.json 和 temp_pdfs 文件夹
          python python_finisher/merge.py

      - name: Upload Final PDF
        uses: actions/upload-artifact@v4
        with:
          name: Wiki-Perfect-PDF
          path: Craft_Engine_Wiki_Perfect.pdf