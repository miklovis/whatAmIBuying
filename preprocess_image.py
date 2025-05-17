from paddleocr import PaddleOCR,draw_ocr
from PIL import Image
import os, re, json, sqlite3
import datetime
from json import JSONEncoder

class Receipt:
    def __init__(self, date, values, amount):
        self.date = str(date)
        self.values = values
        self.amount = str(amount)

class MyEncoder(JSONEncoder):
    def default(self, o):
        return o.__dict__

def crop_receipt(img_path):
    img = Image.open(img_path)
    img = img.crop((0, 965, img.width, img.height))
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
                return int(line[0][2][1])
            
    return find_total_value(img_path, limit + step)

def split_products_and_prices(img_path):
    img = Image.open(img_path)
    

    # img.crop(left px, top px, right px, bot px)
    first_row = img.crop()
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

def divide_by_row(img_path):
    img = Image.open(img_path)
    global amount
    # img.crop(left px, top px, right px, bot px)

    start_px = 0
    step = 54
    row_no = 1

    while (start_px + step) < img.height - step:
        print(start_px + step)
        print(img.height)
        row = img.crop((0, start_px, img.width, start_px + step))
        relative_path = "rows/row.png"
        row.save("./" + relative_path)
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

    for key, value in product_price_dictionary.items():
        try:
            convertedValue = float(value)
            amount = amount + convertedValue
            if convertedValue > 0:
                cleaned_product_price_dictionary[key] = value
        except ValueError:
            pass

    amount = round(amount, 2)

    print(amount)

def find_phrase(phrase, img_path, limit, full_image=False):
    # TODO
    start = 0
    step = 500
    print("LIMIT: ", limit)
    
    img = Image.open(img_path)
    
    if limit >= 1000:
        start = limit - 750

    cropped_image = img.crop((0, start, img.width, limit))
    
    cropped_image.save("cropped_image.png")
    
    result = ocr.ocr("cropped_image.png", cls=False)
    
    for idx in range(len(result)):
        res = result[idx]
        for line in res:
            #print(line[1][0])
            if re.search(phrase, line[1][0]):
                #print(phrase + " FOUND:", line[0][2][1])
                #print(line)
                #print(type(line[0][2][1]))
                return line[1][0]

    return find_phrase(phrase, img_path, limit+500, full_image)

def print_to_file(cleaned_product_price_dictionary, amount):
    date_phrase = find_phrase("Date:", img_path, 500)
    if date_phrase is not None:
        print(date_phrase)
        time_phrase = find_phrase("Time:", img_path, 500)
        if time_phrase is not None:
            print(time_phrase)

    datetime_var = get_datetime_var(date_phrase, time_phrase)
    receipt = Receipt(datetime_var, cleaned_product_price_dictionary, amount)

    # file = open("output.json", "w")
    # file.write(json.dumps(receipt, sort_keys=False))


    with open('output.json', 'w') as f:
        json.dump(receipt, f, cls=MyEncoder, indent=4)

def get_datetime_var(date_string, time_string):
    date_string = str.removeprefix(date_string, "Date: ")
    time_string = str.removeprefix(time_string, "Time: ")

    day, month, year = str.split(date_string, '/')
    hour, minute, second = str.split(time_string, ':')

    year = int("20" + year)
    month = int(month)
    day = int(day)
    hour = int(hour)
    minute = int(minute)
    second = int(second)

    datetime_variable = datetime.datetime(year, month, day, hour, minute, second)
    return datetime_variable

ocr = PaddleOCR(lang='en') # need to run only once to download and load model into memory
img_path = "./IMG_0196.png"
base_path = ""
product_price_dictionary = dict()
cleaned_product_price_dictionary = dict()
amount = 0.0

cropped_image_name = crop_receipt(img_path)

#print(base_path + cropped_image_name)
divide_by_row(base_path + cropped_image_name)

print_to_file(cleaned_product_price_dictionary, amount)


#found_phrase = find_phrase("TOTAL", img_path, 500)
#if found_phrase is not None:
#    print(found_phrase)


