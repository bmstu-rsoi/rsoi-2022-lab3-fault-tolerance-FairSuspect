package models

import (
	"gateway/errors"
	"gateway/objects"
	"gateway/repository"
	"gateway/utils"
)

type TicketsM struct {
	tickets    repository.TicketsRep
	flights    repository.FlightsRep
	privileges repository.PrivilegesRep
}

func NewTicketsM(tickets repository.TicketsRep, flights repository.FlightsRep, privileges repository.PrivilegesRep) *TicketsM {
	return &TicketsM{tickets, flights, privileges}
}

func (model *TicketsM) FetchUser(username string) (*objects.UserInfoResponse, error) {
	data := new(objects.UserInfoResponse)
	tickets, err := model.tickets.GetAll(username)
	if err != nil {
		return nil, err
	}

	flights, err := model.flights.GetAll(1, 100)
	if err == nil {
		data.Tickets = objects.MakeTicketResponseArr(tickets, flights.Items)
	}

	privilege, err := model.privileges.GetAll(username)
	if err == nil {
		data.Privilege = objects.PrivilegeShortInfo{
			Balance: &privilege.Balance,
			Status:  &privilege.Status,
		}
	}
	return data, nil
}

func (model *TicketsM) Fetch() ([]objects.TicketResponse, error) {
	tickets, err := model.tickets.GetAll("")
	if err != nil {
		return nil, err
	}

	flights, err := model.flights.GetAll(1, 100)
	if err != nil {
		utils.Logger.Println("flight service unavaliable")
		flights = &objects.PaginationResponse{}
	}
	return objects.MakeTicketResponseArr(tickets, flights.Items), nil
}

func (model *TicketsM) Create(flight_number string, username string, price int, from_balance bool) (*objects.TicketPurchaseResponse, error) {
	flight, err := model.flights.Find(flight_number)
	if err != nil {
		utils.Logger.Println(err.Error())
		return nil, err
	}

	ticket, err := model.tickets.Create(flight_number, price, username)
	if err != nil {
		utils.Logger.Println(err.Error())
		return nil, err
	}

	privilege, err := model.privileges.Add(username, &objects.AddHistoryRequest{
		TicketUID:       ticket.TicketUid,
		Price:           flight.Price,
		PaidFromBalance: from_balance,
	})
	if err != nil {
		utils.Logger.Println(err.Error())
		model.tickets.Delete(ticket.TicketUid)
		return nil, err
	}

	return objects.NewTicketPurchaseResponse(flight, ticket, privilege), nil
}

func (model *TicketsM) Find(ticketUid string, username string) (*objects.TicketResponse, error) {
	ticket, err := model.tickets.Find(ticketUid)
	if err != nil {
		return nil, err
	} else if username != ticket.Username {
		return nil, errors.ForbiddenTicket
	}

	flight, err := model.flights.Find(ticket.FlightNumber)
	if err != nil {
		utils.Logger.Println("flight service unavaliable")
		flight = &objects.FlightResponse{}
	}
	return objects.ToTicketResponce(ticket, flight), nil
}

func (model *TicketsM) Delete(ticket_uid string, username string) error {
	ticket, err := model.tickets.Find(ticket_uid)
	if err != nil {
		return err
	} else if username != ticket.Username {
		return errors.ForbiddenTicket
	}

	if err = model.tickets.Delete(ticket_uid); err != nil {
		return err
	}

	return model.privileges.Delete(username, ticket_uid)
}
