local function handle()
	local data_string = ctx.cb_data.data
	local parts = {}
	for part in string.gmatch(data_string, "[^|]+") do
	table.insert(parts, part)
	end
	
	local user_id = parts[1]
	local order_id = parts[2]

	local order, error = storage_load("orders", order_id)

	send_keyboard(tonumber(user_id), order.info .. "\n" .. order.phone, {
		Rows = {
		  {
			{Text = "Отказаться", Data = order_id , Script = "declineorder.lua"},
			{Text = "Выполнено", Data = order_id , Script = "doorder.lua"},
		  }
		}})


end

handle()