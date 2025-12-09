def main():
    total = ""
    with open("colors.txt") as f:
        for line in f:
            listItems = line.replace(" ", "").split(",")[:-1]
            for l in listItems[0:len(listItems)//2]:
                item = "color.RGBA{" + l[0:4] + "," + "0x" + l[4:6] + "," + "0x" + l[6:8] + "," + "0x" + l[8:10] + "}, "
                total += item
            total += "\n"
            for l in listItems[len(listItems)//2:]:
                item = "color.RGBA{" + l[0:4] + "," + "0x" + l[4:6] + "," + "0x" + l[6:8] + "," + "0x" + l[8:10] + "}, "
                total += item
            total += "\n"

    print(total)



if __name__ == "__main__":
    main()
