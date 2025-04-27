from paddleocr import PaddleOCR,draw_ocr
ocr = PaddleOCR(lang='en') # need to run only once to download and load model into memory
img_path = "IMG_0196.PNG"
result = ocr.ocr(img_path, cls=False)

with open("imagetext.txt", "w") as file:
    for idx in range(len(result)):
        res = result[idx]
        for line in res:
            file.write(line[1][0] + "\n")
