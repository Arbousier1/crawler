import fitz
import json
import os

def merge():
    # 读取 Go 生成的元数据
    if not os.path.exists("metadata.json"):
        print("❌ 缺失 metadata.json")
        return

    with open("metadata.json", "r", encoding="utf-8") as f:
        metas = json.load(f)
    
    metas.sort(key=lambda x: x['id'])
    doc = fitz.open()
    toc = []
    curr_page = 0
    base_url = "https://xiao-momi.github.io/craft-engine-wiki/"
    temp_dir = "temp_pdfs"

    print(f"📚 正在合并 {len(metas)} 个 PDF 页面并构建目录...")
    for m in metas:
        path = os.path.join(temp_dir, m['path'])
        if not os.path.exists(path): continue
        
        page_doc = fitz.open(path)
        doc.insert_pdf(page_doc)
        
        # 优化标题
        title = m['title'].split('|')[0].split('-')[0].strip()
        
        # 根据 URL 计算层级深度
        rel = m['url'].replace(base_url, "").strip("/")
        level = rel.count("/") + 1 if rel else 1
        
        toc.append([level, title, curr_page + 1])
        curr_page += len(page_doc)
        page_doc.close()

    # --- 核心：修复 Bad Hierarchy Level 错误 ---
    fixed_toc, last_lvl = [], 0
    for l, t, p in toc:
        if l > last_lvl + 1:
            new_lvl = last_lvl + 1
        else:
            new_lvl = l
        fixed_toc.append([new_lvl, t, p])
        last_lvl = new_lvl

    doc.set_toc(fixed_toc)
    doc.save("Craft_Engine_Wiki_Perfect.pdf")
    doc.close()
    print("✨ 最终 PDF 已生成！")

if __name__ == "__main__":
    merge()