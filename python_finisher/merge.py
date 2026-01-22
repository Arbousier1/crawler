import fitz
import json
import os

def merge():
    if not os.path.exists("metadata.json"):
        print("Missing metadata.json")
        return

    with open("metadata.json", "r", encoding="utf-8") as f:
        metas = json.load(f)
    
    metas.sort(key=lambda x: x['id'])
    doc = fitz.open()
    toc = []
    curr_page = 0
    base_url = "https://xiao-momi.github.io/craft-engine-wiki/"
    temp_dir = "temp_pdfs"

    print(f"Merging {len(metas)} pages...")
    for m in metas:
        path = os.path.join(temp_dir, m['path'])
        if not os.path.exists(path): continue
        
        page_doc = fitz.open(path)
        
        # 如果需要彻底删除图片（只留文字），取消下面两行的注释：
        # for page_item in page_doc:
        #     for img in page_item.get_images(): page_doc._deleteObject(img[0])

        doc.insert_pdf(page_doc)
        
        title = m['title'].split('|')[0].split('-')[0].strip()
        rel = m['url'].replace(base_url, "").strip("/")
        level = rel.count("/") + 1 if rel else 1
        toc.append([level, title, curr_page + 1])
        curr_page += len(page_doc)
        page_doc.close()

    # 修正 TOC 层级
    fixed_toc, last_lvl = [], 0
    for l, t, p in toc:
        new_l = last_lvl + 1 if l > last_lvl + 1 else l
        fixed_toc.append([new_l, t, p])
        last_lvl = new_l

    doc.set_toc(fixed_toc)
    
    # 初步压缩保存
    doc.save("Wiki_Raw_Merged.pdf", garbage=4, deflate=True, clean=True)
    doc.close()

if __name__ == "__main__":
    merge()
