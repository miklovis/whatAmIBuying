from paddleocr import PaddleOCR,draw_ocr
from PIL import Image

def crop_receipt(img_path):
    img = Image.open(img_path)
    img = img.crop((0, 950, img.width, img.height))
    img.save("cropped.png")
    
    location = find_total_value(base_path + "cropped.png", 500)
    
    img = img.crop((0, 0, img.width, location))
    img.save("cropped_image.png")
    
    return "cropped_image.png"

def find_total_value(img_path, limit):
    step = 500
    print("LIMIT: ", limit)
    
    img = Image.open(img_path)
    cropped_image = img.crop((0, 0, img.width, limit))
    
    cropped_image.save("cropped_image.png")
    
    result = ocr.ocr("cropped_image.png", cls=False)
    
    for idx in range(len(result)):
        res = result[idx]
        for line in res:
            if line[1][0] == "TOTAL":
                print("TOTAL FOUND:", line[0][2][1])
                print(line)
                print(type(line[0][2][1]))
                return int(line[0][2][1])
            
    return find_total_value(img_path, limit + step)

def split_products_and_prices(img_path):
    img = Image.open(img_path)
    
    product_values = img.crop((0, 0, 850, img.height))
    price_values = img.crop((850, 0, img.width, img.height))
    
    return product_values, price_values

def find_product_names(img):
    img.save("product_names.png")
    img_path = base_path + "product_names.png"
    
    result = ocr.ocr(img_path, cls=False)
    with open("product_names.txt", "w") as file:
        for idx in range(len(result)):
            res = result[idx]
            for line in res:
                file.write(line[1][0] + "\n")
                #print(line)
                
def find_price_values(img):
    img.save("price_values.png")
    img_path = base_path + "price_values.png"
    
    result = ocr.ocr(img_path, cls=False)
    with open("price_values.txt", "w") as file:
        for idx in range(len(result)):
            res = result[idx]
            for line in res:
                file.write(line[1][0] + "\n")
                #print(line)

ocr = PaddleOCR(lang='en') # need to run only once to download and load model into memory
img_path = "C:\\Users\\arnas\\Downloads\\IMG_0196.PNG"
base_path = "C:\\Users\\arnas\\OneDrive\\Documents\\Receipts\\"

cropped_image_name = crop_receipt(img_path)

product_values, price_values = split_products_and_prices(base_path + "cropped_image.png")

find_product_names(product_values)
find_price_values(price_values)

# open product_names.txt and price_values.txt and match the products with their prices
with open("product_names.txt", "r") as product_file:
    product_names = product_file.readlines()
    product_names = [name.strip() for name in product_names]
    
with open("price_values.txt", "r") as price_file:
    price_values = price_file.readlines()
    price_values = [price.strip() for price in price_values]
    
for idx in range(len(product_names)):
    print(product_names[idx], price_values[idx])

"""
finish_flag = False
total_value = 0
with open("output.txt", "w") as file:
    for idx in range(len(result)):
        res = result[idx]
        for line in res:
            file.write(line[1][0] + "\n")
            print(line)
            if finish_flag:
                total_value = line[1][0]
                break
            if line[1][0] == "TOTAL":
                print("TOTAL FOUND")
                finish_flag = True


result = ocr.ocr("C:\\Users\\arnas\\OneDrive\\Documents\\Receipts\\price_values.png", cls=False)
with open("price_values.txt", "w") as file:
    for idx in range(len(result)):
        res = result[idx]
        for line in res:
            file.write(line[1][0] + "\n")
            print(line)
            if(line[1][0] == total_value):
                print("FOUND TOTAL VALUE")
                break
"""
