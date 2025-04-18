from paddleocr import PaddleOCR, draw_ocr
from PIL import Image
import os, re

img_path = "C:\\Users\\AMI01\\Documents\\My Web Sites\\whatAmIBuying\\cropped_image.PNG"
ocr = PaddleOCR(lang='en') # need to run only once to download and load model into memory
base_path = "C:\\Users\\AMI01\\Documents\\My Web Sites\\whatAmIBuying\\"
img = Image.open(img_path)

product_price_dictionary = dict()
# img.crop(left px, top px, right px, bot px)

start_px = 0
step = 54
row_no = 1

while (start_px + step) < img.height - step:
    print(start_px + step)
    print(img.height)
    row = img.crop((0, start_px, img.width, start_px + step))
    relative_path = "rows\\row.png"
    row.save(".\\" + relative_path)
    product = ""
    is_discount = False
    is_collated = False
    is_lacking_letter = False

    result = ocr.ocr(base_path + relative_path, cls=False)
    price_or_letter_string = result[0][len(result[0]) - 1][1][0]
    print("Length of last read string: " + str(len(price_or_letter_string)))
    print("Last read string: " + price_or_letter_string)
    if(price_or_letter_string[0] == '-'):
        is_discount = True
    
    if len(price_or_letter_string) > 5 and is_discount == False and (price_or_letter_string[len(price_or_letter_string) - 1] == 'A' or price_or_letter_string[len(price_or_letter_string) - 1] == 'B'):
        is_collated = True

    if re.search("\\w+\\.\\w+", price_or_letter_string) and is_discount == False and is_collated == False:
        is_lacking_letter = True

    if price_or_letter_string == " A":
        print(is_discount)
        print(is_collated)
        print(is_lacking_letter)

    for idx in range(len(result)):
        res = result[idx]

        if is_discount == True or is_collated == True or is_lacking_letter == True:
            identifier = 1
        else:
            identifier = 2

        for idy in range(len(res) - identifier):
            line = res[idy]
            product += line[1][0] + " "

        for idy in range(len(res)):
            print(res[idy][1][0])

    if is_discount or is_lacking_letter:
        price = result[0][len(result[0]) - 1][1][0]
    elif is_collated:
        price = result[0][len(result[0]) - 1][1][0].split(' ')[0]
    else:
        price = result[0][len(result[0]) - 2][1][0]


    product_price_dictionary[product.strip()] = price

    os.remove(base_path + relative_path)
    start_px += step
    row_no += 1


print(product_price_dictionary)

sum = 0.0

for idx in product_price_dictionary.values():
    try:
        sum += float(idx)
    except ValueError:
        pass


print(round(sum, 2))
