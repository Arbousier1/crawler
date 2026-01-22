import fitz
import json
import os

def merge():
    if not os.path.exists("metadata.json"):
        return

    with open("metadata.json", "r", encoding="utf-8") as f:
        metas = json.load(f)
    
    metas.sort(key=lambda x: x['id'])
    doc = fitz.open()
    toc = []
    curr_page = 0
    base_url = "https://xiao-momi.github.io/craft-engine-wiki/"
    temp_dir = "temp_pdfs"

    for m in metas:
        path = os.path.join(temp_dir, m['path'])
        if not os.path.exists(path): continue
        
        page_doc = fitz.open(path)
        # --- AI 优化：删除所有图像（如果 AI 只需要文本） ---
        # 如果你需要保留图表请注释掉下面这行
        # for page in page_doc:
        #    for img in page.get_images(): page_doc._deleteObject(img[0])
            
        doc.insert_pdf(page_doc)
        
        clean_title = m['title'].split('|')[0].split('-')[0].strip()
        rel = m['url'].replace(base_url, "").strip("/")
        level = rel.count("/") + 1 if rel else 1
        
        toc.append([level, clean_title, curr_page + 1])
        curr_page += len(page_doc)
        page_doc.close()

    fixed_toc, last_lvl = [], 0
    for l, t, p in toc:
        new_l = last_lvl + 1 if l > last_lvl + 1 else l
        fixed_toc.append([new_l, t, p])
        last_lvl = new_l

    doc.set_toc(fixed_toc)
    
    # --- 极限压缩保存设置 ---
    # garbage=4: 彻底清理未使用对象
    # deflate=True: 压缩流
    # clean=True: 清理内容流
    output_name = "Wiki_AI_Raw.pdf"
    doc.save(output_name, garbage=4, deflate=True, clean=True)
    doc.close()
    print("✨ 初步压缩完成。")

if __name__ == "__main__":
    merge()
