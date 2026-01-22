import fitz, json, os

def merge():
    with open("metadata.json", "r", encoding="utf-8") as f:
        metas = json.load(f)
    
    # 按 ID 排序
    metas.sort(key=lambda x: x['id'])
    
    doc = fitz.open()
    toc = []
    curr_page = 0
    base_url = "https://xiao-momi.github.io/craft-engine-wiki/"

    for m in metas:
        page_doc = fitz.open(m['path'])
        doc.insert_pdf(page_doc)
        
        # 计算层级
        rel = m['url'].replace(base_url, "").strip("/")
        level = rel.count("/") + 1 if rel else 1
        
        toc.append([level, m['title'], curr_page + 1])
        curr_page += len(page_doc)
        page_doc.close()

    # 修正 TOC (防止 bad hierarchy level)
    fixed_toc, last_lvl = [], 0
    for l, t, p in toc:
        new_l = last_lvl + 1 if l > last_lvl + 1 else l
        fixed_toc.append([new_l, t, p])
        last_lvl = new_l

    doc.set_toc(fixed_toc)
    doc.save("Final_Wiki_Hybrid.pdf")
    print("PDF Merged Successfully with TOC!")

if __name__ == "__main__":
    merge()
